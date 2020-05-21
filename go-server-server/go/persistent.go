package mseeserver

import (
    "fmt"
    "github.com/go-redis/redis/v7"
    "log"
    "strconv"
    "strings"
    "swsscommon"
    "time"
    "bytes"
    "github.com/satori/go.uuid"
)

const ServerAPIVersion string = "1.0.0"
var ServerResetGuid string
var ServerResetTime string

var redisDB *redis.Client
var swssDB swsscommon.DBConnector
var swss_conf_DB swsscommon.DBConnector
var trustedertCommonNames []string

var vnetGuidMap map[string]uint32
var vnetGuidIdUsed []bool
var nextGuidId uint32

const REDIS_SOCK string = "/var/run/redis/redis.sock"

const APPL_DB int = 0
const COUNTER_DB int = 2
const CONFIG_DB int = 4

// TODO:
// DB 4 is reserved for config DB, we can not simply flush it any more
// when we reset the server. It will affect other applications on Sonic.
// Let use use DB 7 for caching now and move the info in cache to config DB
// with new schema and delete only keys for this server in Reset.
const APPL_CACHE_DB int = 7

const SWSS_TIMEOUT uint = 0

// DB Table names
const VXLAN_TUNNEL_TB   string = "VXLAN_TUNNEL"
const VNET_TB           string = "VNET"
const VLAN_TB           string = "VLAN"
const VLAN_INTF_TB      string = "VLAN_INTERFACE"
const VLAN_MEMB_TB      string = "VLAN_MEMBER"
const VLAN_NEIGH_TB     string = "NEIGH"
const ROUTE_TUN_TB      string = "VNET_ROUTE_TUNNEL_TABLE"
const LOCAL_ROUTE_TB    string = "VNET_ROUTE_TABLE"
const CFG_ROUTE_TUN_TB  string = "VNET_ROUTE_TUNNEL"
const CFG_LOCAL_ROUTE_TB    string = "VNET_ROUTE"

// DB Helper constants
const VNET_NAME_PREF  string = "Vnet"
const VLAN_NAME_PREF  string = "Vlan"

type db_ops struct {
   separator string
   swss_db  swsscommon.DBConnector
   db_num   int
}

var app_db_ops db_ops
var conf_db_ops db_ops

func Initialise() {
    DBConnect()
    InitialiseVariables()
}

func InitialiseVariables() {
    trustedertCommonNames = strings.Split(*ClientCertCommonNameFlag, ",")
    var err error
    ServerResetGuid, ServerResetTime, err = CacheGetConfigResetInfo()

    if err == redis.Nil {
        loc, _ := time.LoadLocation("UTC")
        now := time.Now().In(loc)
        ServerResetTime = fmt.Sprintf("%v", now)

        newuuid := uuid.NewV4()
        ServerResetGuid = newuuid.String()

        err = CacheSetConfigResetInfo(ServerResetGuid, ServerResetTime)
        if err != nil {
            log.Fatalf("error: could not save reset info to DB, error: %+v", err)
        }
        log.Printf("info: set config reset Guid and Time to %v, %v", ServerResetGuid, ServerResetTime)
    } else if err == nil {
        log.Printf("info: find config reset Guid and Time in DB as %v, %v", ServerResetGuid, ServerResetTime)
    } else {
        log.Fatalf("error: could not retrieve server reset info from DB, error: %+v", err)
    }
    genVnetGuidMap()
}


func genVnetGuidMap() {
    vnetGuidMap = make(map[string]uint32)
    var max_ind, i uint32 = 0, 0
    db := &conf_db_ops
    kv, err := GetKVsMulti(db.db_num, generateDBTableKey(db.separator, VNET_TB, "*"))

    if (err != nil) || (len(kv) == 0) {
        log.Printf("info: No VNET tables found to gen Vnet Guid Map, default init, err: %d len kv: %d", err, len(kv))
        vnetGuidIdUsed = make([]bool, 0, 30) /* 30 HSM/rack? */
        nextGuidId = 1
        return
    }

    for k, v := range kv {
         log.Printf("info: TABLE: %s TABLE_KVs: %#v", k, v)
         tb_key_len := len(generateDBTableKey(db.separator, VNET_TB, VNET_NAME_PREF))
         vnet_id_str := k[tb_key_len:]
         vnet_id_64, err_c := strconv.ParseUint(vnet_id_str, 10, 32)
         if (err_c != nil) || (vnet_id_64 == 0) {
             log.Printf("error: Found non integer vnet_id %s", vnet_id_str)
         }
         vnet_id := uint32(vnet_id_64)
         if v["guid"] == "" {
             log.Printf("error: Found nil guid %s %s", k, vnet_id_str)
             continue
         }
         log.Printf("info: storing vnet-guid: %s, vnet_id %d", v["guid"], vnet_id)
         vnetGuidMap[v["guid"]] = vnet_id
         if vnet_id > max_ind {
             max_ind = vnet_id
         }
    }
    vnetGuidIdUsed = make([]bool, max_ind, max_ind)
    for _, v := range vnetGuidMap {
        vnetGuidIdUsed[v - 1] = true
    }
    for i = 0; i < max_ind; i++ {
        if !vnetGuidIdUsed[i] {
             break
        }
    }
    nextGuidId = i + 1
}

func DBConnect() {
    if *RunApiAsLocalTestDocker {
        redisDB = redis.NewClient(&redis.Options{
            Addr:     "localhost:6379",
            Password: "",
        })
    } else {
        redisDB = redis.NewClient(&redis.Options{
            Network:  "unix",
            Addr:     REDIS_SOCK,
            Password: "",
        })
    }

    log.Printf("info: Redis connection established (%+v)", redisDB)

    if *RunApiAsLocalTestDocker {
	     swssDB = swsscommon.NewDBConnector(APPL_DB, "localhost", 6379, SWSS_TIMEOUT)
	     swss_conf_DB = swsscommon.NewDBConnector(CONFIG_DB, "localhost", 6379, SWSS_TIMEOUT)
    } else {
        swssDB = swsscommon.NewDBConnector2(APPL_DB, REDIS_SOCK, SWSS_TIMEOUT)
	     swss_conf_DB = swsscommon.NewDBConnector2(CONFIG_DB, REDIS_SOCK, SWSS_TIMEOUT)
    }
    app_db_ops = db_ops{separator: ":", swss_db: swssDB, db_num: APPL_DB}
    conf_db_ops = db_ops{separator: "|", swss_db: swss_conf_DB, db_num: CONFIG_DB}
}

func GetKVs(DB int, key string) (kv map[string]string, err error) {
    pipe := redisDB.TxPipeline()
    pipe.Select(DB)
    kvRes := pipe.HGetAll(key)
    _, err = pipe.Exec()
    if err != nil {
        return
    }

    kv = kvRes.Val()
    if len(kv) == 0 {
        kv = nil
    }

    return
}

func GetKVsMulti(DB int, pattern string) (kv map[string]map[string]string, err error) {
    var cursor uint64

    kv = make(map[string]map[string]string)

    for {
        pipe := redisDB.TxPipeline()
        pipe.Select(DB)
        ret := pipe.Scan(cursor, pattern, 1)

        _, err = pipe.Exec()
        if err != nil {
            return
        }

        var keys []string
        keys, cursor = ret.Val()

        for _, key := range keys {
            kv[key], err = GetKVs(DB, key)
            if err != nil {
                return
            }
        }

        if cursor == 0 {
            break
        }
    }

    return
}

func SwssGetVrouterRoutes(vnet_id_str string, vnidMatch int, ipFilter string) (routes []RouteModel, err error) {
    db := &conf_db_ops
    var pattern string
    // TODO: Keep only else statement code in production
    pattern = generateDBTableKey(db.separator, CFG_ROUTE_TUN_TB, vnet_id_str, ipFilter)
    routes = []RouteModel{}

    kv, err := GetKVsMulti(db.db_num,pattern)
    if err != nil {
        return
    }

    for k, kvp := range kv {
        ipprefix := strings.Split(k, db.separator)[2]

        routeModel := RouteModel{
            IPPrefix:    ipprefix,
            NextHop:     kvp["endpoint"],
        }

        if vnid, ok := kvp["vni"]; ok {
            routeModel.Vnid, _ = strconv.Atoi(vnid)
        }

        if vnidMatch >= 0 {
            if vnidMatch != routeModel.Vnid {
                continue
            }
        }

        if mac, ok := kvp["mac_address"]; ok {
            routeModel.MACAddress = mac
        }

        routes = append(routes, routeModel)
    }

    return
}

func CacheGetConfigResetInfo() (GUID string, time string, err error) {
    pipe := redisDB.TxPipeline()
    pipe.Select(APPL_CACHE_DB)
    getCmd_guid := pipe.HGet("RESET_INFO", "GUID")
    getCmd_time := pipe.HGet("RESET_INFO", "time")
    _, err = pipe.Exec()

    if err != nil {
        return
    }

    if getCmd_guid.Err() != nil {
        err = getCmd_guid.Err()
        return
    }

    if getCmd_time.Err() != nil {
        err = getCmd_time.Err()
        return
    }

    return getCmd_guid.Val(), getCmd_time.Val(), nil
}

func CacheSetConfigResetInfo(GUID string, time string) error {
    pipe := redisDB.TxPipeline()
    pipe.Select(APPL_CACHE_DB)
    setCmd_guid := pipe.HSet("RESET_INFO", "GUID", GUID)
    setCmd_time := pipe.HSet("RESET_INFO", "time", time)
    _, err := pipe.Exec()
    if err != nil {
        return err
    }

    if setCmd_guid.Err() != nil {
        return setCmd_guid.Err()
    }

    if setCmd_time.Err() != nil {
        return setCmd_time.Err()
    }

    return nil
}

func CacheGetVnetGuidId(GUID string) (val uint32) {
    val = vnetGuidMap[GUID]
    return
}

func CacheGenAndSetVnetGuidId(GUID string) (val uint32) {
    vnetGuidMap[GUID] = nextGuidId
    val = nextGuidId
    if nextGuidId == (uint32)(len(vnetGuidIdUsed) + 1) {
        vnetGuidIdUsed = append(vnetGuidIdUsed, true)
        nextGuidId++
    } else {
        var i uint32
        vnetGuidIdUsed[nextGuidId - 1] = true
        for i = nextGuidId; i < (uint32) (len(vnetGuidIdUsed)); i++ {
            if !vnetGuidIdUsed[i] {
                break
            }
        }
        nextGuidId = i + 1
    }
    return
}

func CacheDeleteVnetGuidId(GUID string) {
    i := vnetGuidMap[GUID]
    vnetGuidIdUsed[i - 1] = false
    if i < nextGuidId {
         nextGuidId = i
    }
    delete(vnetGuidMap, GUID)
}

func generateDBTableKey(separator string, vars ...string) (string) {
     var buf bytes.Buffer
     for i := 0; i < len(vars) ; i++ {
         buf.WriteString(vars[i])
         if i != (len(vars) - 1) {
               buf.WriteString(separator)
         }
    }
    return buf.String()
}
