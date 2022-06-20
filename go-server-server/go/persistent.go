package restapi

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

var ConfigResetStatus bool

var redisDB *redis.Client
var swssDB swsscommon.DBConnector
var swss_conf_DB swsscommon.DBConnector
var swss_ctr_DB swsscommon.DBConnector
var trustedCertCommonNames []string

var vnetGuidMap map[string]uint32
var vniVnetMap map[uint32]string
var vnetGuidIdUsed []bool
var nextGuidId uint32
var localTunnelLpbkIps []string

const REDIS_SOCK string = "/var/run/redis/redis.sock"

const APPL_DB int = 0
const COUNTER_DB int = 2
const CONFIG_DB int = 4

// Use RESTAPI_DB for cache
const APPL_CACHE_DB int = 8

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
const CRM_TB            string = "CRM"
const STATIC_ROUTE_TB   string = "STATIC_ROUTE"
const BGP_PROFILE_TABLE       string = "BGP_PROFILE_TABLE"

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
var ctr_db_ops db_ops

func Initialise() {
    DBConnect()
    InitialiseVariables()
}

func InitialiseVariables() {
    trustedCertCommonNames = strings.Split(*ClientCertCommonNameFlag, ",")
    var err error
    var resetStatus string
    ServerResetGuid, ServerResetTime, resetStatus, err = CacheGetConfigResetInfo()

    if err == redis.Nil {
        loc, _ := time.LoadLocation("UTC")
        now := time.Now().In(loc)
        ServerResetTime = fmt.Sprintf("%v", now)

        newuuid,_ := uuid.NewV4()
        ServerResetGuid = newuuid.String()

        err = CacheSetConfigResetInfo(ServerResetGuid, ServerResetTime)
        if err != nil {
            log.Fatalf("error: could not save reset info to DB, error: %+v", err)
        }
        log.Printf("info: set config reset Guid and Time to %v, %v", ServerResetGuid, ServerResetTime)

        ConfigResetStatus = true
        err = CacheSetResetStatusInfo(ConfigResetStatus)
        if err != nil {
            log.Fatalf("error: could not save reset status info to DB, error: %+v", err)
        }
    } else if err == nil {
        ConfigResetStatus = (resetStatus == "true")
        log.Printf("info: find config reset Guid, Time, ResetStatus in DB as %v, %v, %v",
                   ServerResetGuid, ServerResetTime, ConfigResetStatus)
    } else {
        log.Fatalf("error: could not retrieve server reset info from DB, error: %+v", err)
    }
    genVnetGuidMap()

    genVxlanTunnelInfo()
}

func genVnetGuidMap() {
    vnetGuidMap = make(map[string]uint32)
    vniVnetMap = make(map[uint32]string)
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
         if v["vni"] == "" {
            log.Printf("error: Found nil vni %s %s", k, vnet_id_str)
            continue
        }
         log.Printf("info: storing vnet-guid: %s, vnet_id %d", v["guid"], vnet_id)
         vnetGuidMap[v["guid"]] = vnet_id
         vni, _ := strconv.Atoi(v["vni"])
         vniVnetMap[uint32(vni)] = v["guid"]
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

func genVxlanTunnelInfo() {
    localTunnelLpbkIps = make([]string, 256)
    db := &conf_db_ops
    kv, err := GetKVsMulti(db.db_num, generateDBTableKey(db.separator, VXLAN_TUNNEL_TB, "*"))

    if (err != nil) || (len(kv) == 0) {
        log.Printf("info: No Vxlan tunnel tables, default init, err: %d len kv: %d", err, len(kv))
        return
    }

    for k, v := range kv {
        log.Printf("info: TABLE: %s TABLE_KVs: %#v", k, v)
        localTunnelLpbkIps = append(localTunnelLpbkIps, v["src_ip"])
    }

    log.Printf("info: Existing loopback ips %v", localTunnelLpbkIps)
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
        swss_ctr_DB = swsscommon.NewDBConnector2(COUNTER_DB, REDIS_SOCK, SWSS_TIMEOUT)
    }
    app_db_ops = db_ops{separator: ":", swss_db: swssDB, db_num: APPL_DB}
    conf_db_ops = db_ops{separator: "|", swss_db: swss_conf_DB, db_num: CONFIG_DB}
    ctr_db_ops = db_ops{separator: ":", swss_db: swss_ctr_DB, db_num: COUNTER_DB}
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
    db := &app_db_ops
    var pattern string

    rt_tb_name := ROUTE_TUN_TB
    if *RunApiAsLocalTestDocker {
        rt_tb_name = "_"+ROUTE_TUN_TB
    }
    pattern = generateDBTableKey(db.separator, rt_tb_name, vnet_id_str, ipFilter)
    routes = []RouteModel{}

    kv1, err1 := GetKVsMulti(db.db_num,pattern)
    if err1 != nil {
        return
    }

    rt_tb_name = LOCAL_ROUTE_TB
    if *RunApiAsLocalTestDocker {
        rt_tb_name = "_"+LOCAL_ROUTE_TB
    }
    pattern = generateDBTableKey(db.separator, rt_tb_name, vnet_id_str, ipFilter)

    kv2, err2 := GetKVsMulti(db.db_num,pattern)
    if err2 != nil {
        return
    }        

    for k, kvp := range kv1 {
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

        if weight, ok := kvp["weight"]; ok {
            routeModel.Weight = weight
        }

        if profile, ok := kvp["profile"]; ok {
            routeModel.Profile = profile
        }

        routes = append(routes, routeModel)
    }

    for k, kvp := range kv2 {
        ipprefix := strings.Split(k, db.separator)[2]

        routeModel := RouteModel{
            IPPrefix:    ipprefix,
            NextHop:     kvp["nexthop"],
        }

        if ifname, ok := kvp["ifname"]; ok {
            routeModel.IfName = ifname
        }

        if weight, ok := kvp["weight"]; ok {
            routeModel.Weight = weight
        }

        if profile, ok := kvp["profile"]; ok {
            routeModel.Profile = profile
        }
        routes = append(routes, routeModel)
    }
    return
}

func CacheGetConfigResetInfo() (GUID string, time string, resetStatus string, err error) {
    pipe := redisDB.TxPipeline()
    pipe.Select(APPL_CACHE_DB)
    getCmd_guid := pipe.HGet("RESET_INFO", "GUID")
    getCmd_time := pipe.HGet("RESET_INFO", "time")
    getCmd_resetStatus := pipe.HGet("RESET_INFO", "reset_status")
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

    if getCmd_resetStatus.Err() != nil {
        err = getCmd_resetStatus.Err()
        return
    }

    return getCmd_guid.Val(), getCmd_time.Val(), getCmd_resetStatus.Val(), nil
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

func CacheSetResetStatusInfo(resetStatus bool) error {
    pipe := redisDB.TxPipeline()
    pipe.Select(APPL_CACHE_DB)

    val := "false"
    if resetStatus {
        val = "true"
    }

    setCmd_resetStatus := pipe.HSet("RESET_INFO", "reset_status", val)
    _, err := pipe.Exec()
    if err != nil {
        return err
    }

    if setCmd_resetStatus.Err() != nil {
        return setCmd_resetStatus.Err()
    }

    return nil
}

func CacheTunnelLpbkIps(ipAddr string, add bool) {

    log.Printf("info: lbkp ip update %s, add: %v", ipAddr, add)

	if add {
	    localTunnelLpbkIps = append(localTunnelLpbkIps, ipAddr)
	    log.Printf("info: stored loopback ips %v", localTunnelLpbkIps)
	} else {
        //Deleting tunnel is not currently supported.
	}
}

func CacheGetVnetGuidId(GUID string) (val uint32) {
    val = vnetGuidMap[GUID]
    return
}

func CacheGetVniId(VNI uint32) (val string) {
    val = vniVnetMap[VNI]
    return
}

func CacheGenAndSetVnetGuidId(GUID string, VNI uint32) (val uint32) {
    vnetGuidMap[GUID] = nextGuidId
    vniVnetMap[VNI] = GUID
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
    for k, v := range vniVnetMap {
        if v == GUID {
            delete(vniVnetMap, k)
            break
        }
    }
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
