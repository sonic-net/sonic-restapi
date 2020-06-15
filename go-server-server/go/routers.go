package restapi

import (
    "net/http"
    "fmt"
    "log"
    "time"
    "sync"
    "github.com/gorilla/mux"
)

type Route struct {
    Name        string
    Method      string
    Pattern     string
    HandlerFunc http.HandlerFunc
}

type Routes []Route

var writeMutex sync.Mutex

func Middleware(inner http.Handler, name string) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        log.Printf(
            "info: request: %s %s %s",
            r.Method,
            r.RequestURI,
            name,
        )

        if r.TLS == nil || CommonNameMatch(r) {
            log.Printf("trace: acquire server write lock")
            writeMutex.Lock()
            inner.ServeHTTP(NewLoggingResponseWriter(w), r)
            writeMutex.Unlock()
            log.Printf("trace: release server write lock")
        } else {
            WriteRequestError(NewLoggingResponseWriter(w), http.StatusUnauthorized,
                        "Authentication Fail with untrusted client cert", []string{}, "")
        }

        log.Printf(
            "info: request: duration %s",
            time.Since(start),
        )
    })
}

func NewRouter() *mux.Router {
    router := mux.NewRouter().StrictSlash(true)
    for _, route := range routes {
        handler := Middleware(route.HandlerFunc, route.Name)

        router.
            Methods(route.Method).
            Path(route.Pattern).
            Name(route.Name).
            Handler(handler)
    }

    return router
}

func Index(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Sonic MSEE Restful API v1!")
}

var routes = Routes{
    Route{
        "Index",
        "GET",
        "/v1/",
        Index,
    },

    Route{
        "StateHeartbeatGet",
        "GET",
        "/v1/state/heartbeat",
        StateHeartbeatGet,
    },

    Route{
        "ConfigResetStatusGet",
        "GET",
        "/v1/config/resetstatus",
        ConfigResetStatusGet,
    },

    Route{
        "ConfigResetStatusPost",
        "POST",
        "/v1/config/resetstatus",
        ConfigResetStatusPost,
    },

    Route{
        "ConfigInterfaceVlanDelete",
        "DELETE",
        "/v1/config/interface/vlan/{vlan_id}",
        ConfigInterfaceVlanDelete,
    },

    Route{
        "ConfigInterfaceVlanGet",
        "GET",
        "/v1/config/interface/vlan/{vlan_id}",
        ConfigInterfaceVlanGet,
    },

    Route{
        "ConfigInterfaceVlanPost",
        "POST",
        "/v1/config/interface/vlan/{vlan_id}",
        ConfigInterfaceVlanPost,
    },

    Route{
        "ConfigInterfaceVlansGet",
        "GET",
        "/v1/config/interface/vlans",
        ConfigInterfaceVlansGet,
    },

    Route{
        "ConfigInterfaceVlansAllGet",
        "GET",
        "/v1/config/interface/vlans/all",
        ConfigInterfaceVlansAllGet,
    },

    Route{
        "ConfigInterfaceVlanMemberDelete",
        "DELETE",
        "/v1/config/interface/vlan/{vlan_id}/member/{if_name}",
        ConfigInterfaceVlanMemberDelete,
    },

    Route{
        "ConfigInterfaceVlanMemberGet",
        "GET",
        "/v1/config/interface/vlan/{vlan_id}/member/{if_name}",
        ConfigInterfaceVlanMemberGet,
    },

    Route{
        "ConfigInterfaceVlanMemberPost",
        "POST",
        "/v1/config/interface/vlan/{vlan_id}/member/{if_name}",
        ConfigInterfaceVlanMemberPost,
    },

    Route{
        "ConfigInterfaceVlanMembersGet",
        "GET",
        "/v1/config/interface/vlan/{vlan_id}/members",
        ConfigInterfaceVlanMembersGet,
    },

    Route{
        "ConfigInterfaceVlanNeighborDelete",
        "DELETE",
        "/v1/config/interface/vlan/{vlan_id}/neighbor/{ip_addr}",
        ConfigInterfaceVlanNeighborDelete,
    },

    Route{
        "ConfigInterfaceVlanNeighborGet",
        "GET",
        "/v1/config/interface/vlan/{vlan_id}/neighbor/{ip_addr}",
        ConfigInterfaceVlanNeighborGet,
    },

    Route{
        "ConfigInterfaceVlanNeighborPost",
        "POST",
        "/v1/config/interface/vlan/{vlan_id}/neighbor/{ip_addr}",
        ConfigInterfaceVlanNeighborPost,
    },

    Route{
        "ConfigInterfaceVlanNeighborsGet",
        "GET",
        "/v1/config/interface/vlan/{vlan_id}/neighbors",
        ConfigInterfaceVlanNeighborsGet,
    },

    Route{
        "ConfigTunnelDecapTunnelTypeDelete",
        "DELETE",
        "/v1/config/tunnel/decap/{tunnel_type}",
        ConfigTunnelDecapTunnelTypeDelete,
    },

    Route{
        "ConfigTunnelDecapTunnelTypeGet",
        "GET",
        "/v1/config/tunnel/decap/{tunnel_type}",
        ConfigTunnelDecapTunnelTypeGet,
    },

    Route{
        "ConfigTunnelDecapTunnelTypePost",
        "POST",
        "/v1/config/tunnel/decap/{tunnel_type}",
        ConfigTunnelDecapTunnelTypePost,
    },

    Route{
        "ConfigTunnelEncapVxlanVnidDelete",
        "DELETE",
        "/v1/config/tunnel/encap/vxlan/{vnid}",
        ConfigTunnelEncapVxlanVnidDelete,
    },

    Route{
        "ConfigTunnelEncapVxlanVnidGet",
        "GET",
        "/v1/config/tunnel/encap/vxlan/{vnid}",
        ConfigTunnelEncapVxlanVnidGet,
    },

    Route{
        "ConfigTunnelEncapVxlanVnidPost",
        "POST",
        "/v1/config/tunnel/encap/vxlan/{vnid}",
        ConfigTunnelEncapVxlanVnidPost,
    },

    Route{
        "ConfigVrouterVrfIdDelete",
        "DELETE",
        "/v1/config/vrouter/{vnet_name}",
        ConfigVrouterVrfIdDelete,
    },

    Route{
        "ConfigVrouterVrfIdGet",
        "GET",
        "/v1/config/vrouter/{vnet_name}",
        ConfigVrouterVrfIdGet,
    },

    Route{
        "ConfigVrouterVrfIdPost",
        "POST",
        "/v1/config/vrouter/{vnet_name}",
        ConfigVrouterVrfIdPost,
    },

    Route{
        "ConfigVrouterVrfIdRoutesDelete",
        "DELETE",
        "/v1/config/vrouter/{vnet_name}/routes",
        ConfigVrouterVrfIdRoutesDelete,
    },

    Route{
        "ConfigVrouterVrfIdRoutesGet",
        "GET",
        "/v1/config/vrouter/{vnet_name}/routes",
        ConfigVrouterVrfIdRoutesGet,
    },

    Route{
        "ConfigVrouterVrfIdRoutesPatch",
        "PATCH",
        "/v1/config/vrouter/{vnet_name}/routes",
        ConfigVrouterVrfIdRoutesPatch,
    },

    Route{
        "StateInterfacePortGet",
        "GET",
        "/v1/state/interface/{port}",
        StateInterfacePortGet,
    },

    Route{
        "StateInterfaceGet",
        "GET",
        "/v1/state/interface",
        StateInterfaceGet,
    },

    // Required to run Unit tests
    Route{
        "InMemConfigRestart",
        "POST",
        "/v1/config/restartdb",
        InMemConfigRestart,
    },

   // Adding Ping method from VRF context
   Route{
        "Ping",
        "POST",
        "/v1/operations/ping",
        Ping,
    },

}
