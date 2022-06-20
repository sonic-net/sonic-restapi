import json
import logging
import requests

# DB Names
ROUTE_TUN_TB      = "_VNET_ROUTE_TUNNEL_TABLE"
LOCAL_ROUTE_TB    = "_VNET_ROUTE_TABLE"

# DB Helper constants
VNET_NAME_PREF    = "Vnet"
VLAN_NAME_PREF    = "Vlan"

TEST_HOST = 'http://localhost:8090/'

class RESTAPI_client:

    def __init__(self, db):
        self.db = db

    maxDiff = None

    def post(self, url, body = []):
        if body == None:
            data = None
        else:
            data = json.dumps(body)

        logging.info("Request POST: %s" % url)
        logging.info("JSON Body: %s" % data)
        r = requests.post(TEST_HOST + url, data=data, headers={'Content-Type': 'application/json'})
        logging.info('Response Code: %s' % r.status_code)
        logging.info('Response Body: %s' % r.text)
        return r

    def patch(self, url, body = []):
        if body == None:
            data = None
        else:
            data = json.dumps(body)

        logging.info("Request PATCH: %s" % url)
        logging.info("JSON Body: %s" % data)
        r = requests.patch(TEST_HOST + url, data=data, headers={'Content-Type': 'application/json'})
        logging.info('Response Code: %s' % r.status_code)
        logging.info('Response Body: %s' % r.text)
        return r

    def get(self, url, body = [], params = {}):
        if body == None:
            data = None
        else:
            data = json.dumps(body)

        logging.info("Request GET: %s" % url)
        logging.info("JSON Body: %s" % data)
        r = requests.get(TEST_HOST + url, data=data, params=params, headers={'Content-Type': 'application/json'})
        logging.info('Response Code: %s' % r.status_code)
        logging.info('Response Body: %s' % r.text)
        return r

    def delete(self, url, body = [], params = {}):
        if body == None:
            data = None
        else:
            data = json.dumps(body)

        logging.info("Request DELETE: %s" % url)
        logging.info("JSON Body: %s" % data)
        r = requests.delete(TEST_HOST + url, data=data, params=params, headers={'Content-Type': 'application/json'})
        logging.info('Response Code: %s' % r.status_code)
        logging.info('Response Body: %s' % r.text)
        return r

    def get_config_reset_status(self):
        return self.get('v1/config/resetstatus')

    def post_config_reset_status(self, value):
        return self.post('v1/config/resetstatus', value)

    # BGP Community String
    def get_bgp_community_string(self, profile_name):
        return self.get('v1/config/bgp/profile/{profile_name}'.format(profile_name=profile_name))

    def post_bgp_community_string(self, profile_name, value):
        return self.post('v1/config/bgp/profile/{profile_name}'.format(profile_name=profile_name), value)

    def delete_bgp_community_string(self, profile_name):
        return self.delete('v1/config/bgp/profile/{profile_name}'.format(profile_name=profile_name))

    # VRF/VNET
    def post_config_vrouter_vrf_id(self, vrf_id, value):
        return self.post('v1/config/vrouter/{vrf_id}'.format(vrf_id=vrf_id), value)

    def get_config_vrouter_vrf_id(self, vrf_id):
        return self.get('v1/config/vrouter/{vrf_id}'.format(vrf_id=vrf_id))

    def delete_config_vrouter_vrf_id(self, vrf_id):
        return self.delete('v1/config/vrouter/{vrf_id}'.format(vrf_id=vrf_id))

    # Encap
    def post_config_tunnel_encap_vxlan_vnid(self, vnid, value):
        return self.post('v1/config/tunnel/encap/vxlan/{vnid}'.format(vnid=vnid), value)

    def delete_config_tunnel_encap_vxlan_vnid(self, vnid):
        return self.delete('v1/config/tunnel/encap/vxlan/{vnid}'.format(vnid=vnid))

    def get_config_tunnel_encap_vxlan_vnid(self, vnid):
        return self.get('v1/config/tunnel/encap/vxlan/{vnid}'.format(vnid=vnid))

    # Decap
    def post_config_tunnel_decap_tunnel_type(self, tunnel_type, value):
        return self.post('v1/config/tunnel/decap/{tunnel_type}'.format(tunnel_type=tunnel_type), value)

    def get_config_tunnel_decap_tunnel_type(self, tunnel_type):
        return self.get('v1/config/tunnel/decap/{tunnel_type}'.format(tunnel_type=tunnel_type))

    def delete_config_tunnel_decap_tunnel_type(self, tunnel_type):
        return self.delete('v1/config/tunnel/decap/{tunnel_type}'.format(tunnel_type=tunnel_type))

    # Vlan
    def post_config_vlan(self, vlan_id, value):
        return self.post('v1/config/interface/vlan/{vlan_id}'.format(vlan_id=vlan_id), value) 

    def get_config_vlan(self, vlan_id):
        return self.get('v1/config/interface/vlan/{vlan_id}'.format(vlan_id=vlan_id)) 

    def delete_config_vlan(self, vlan_id):
        return self.delete('v1/config/interface/vlan/{vlan_id}'.format(vlan_id=vlan_id)) 

    def get_config_interface_vlans(self, vnet_id=None):
        params = {}
        if vnet_id != None:
            params['vnet_id'] = vnet_id
        return self.get('v1/config/interface/vlans',params=params)

    def get_config_vlans_all(self):
        return self.get('v1/config/interface/vlans/all')

    # Vlan Member
    def post_config_vlan_member(self, vlan_id, if_name, value):
        return self.post('v1/config/interface/vlan/{vlan_id}/member/{if_name}'.format(vlan_id=vlan_id, if_name=if_name), value) 

    def get_config_vlan_member(self, vlan_id, if_name):
        return self.get('v1/config/interface/vlan/{vlan_id}/member/{if_name}'.format(vlan_id=vlan_id, if_name=if_name)) 

    def delete_config_vlan_member(self, vlan_id, if_name):
        return self.delete('v1/config/interface/vlan/{vlan_id}/member/{if_name}'.format(vlan_id=vlan_id, if_name=if_name)) 

    def get_config_interface_vlan_members(self, vlan_id):
        return self.get('v1/config/interface/vlan/{vlan_id}/members'.format(vlan_id=vlan_id))

    def get_config_members_all(self):
        return self.get('v1/config/interface/vlans/members/all')

    # Vlan Neighbor
    def post_config_vlan_neighbor(self, vlan_id, ip_addr):
        return self.post('v1/config/interface/vlan/{vlan_id}/neighbor/{ip_addr}'.format(vlan_id=vlan_id, ip_addr=ip_addr)) 

    def get_config_vlan_neighbor(self, vlan_id, ip_addr):
        return self.get('v1/config/interface/vlan/{vlan_id}/neighbor/{ip_addr}'.format(vlan_id=vlan_id, ip_addr=ip_addr)) 

    def delete_config_vlan_neighbor(self, vlan_id, ip_addr):
        return self.delete('v1/config/interface/vlan/{vlan_id}/neighbor/{ip_addr}'.format(vlan_id=vlan_id, ip_addr=ip_addr))    

    def get_config_interface_vlan_neighbors(self, vlan_id):
        return self.get('v1/config/interface/vlan/{vlan_id}/neighbors'.format(vlan_id=vlan_id))

    # Routes
    def patch_config_vrouter_vrf_id_routes(self, vrf_id, value):
        return self.patch('v1/config/vrouter/{vrf_id}/routes'.format(vrf_id=vrf_id), value)

    def patch_config_vrf_vrf_id_routes(self, vrf_id, value):
        return self.patch('v1/config/vrf/{vrf_id}/routes'.format(vrf_id=vrf_id), value)

    def delete_config_vrouter_vrf_id_routes(self, vrf_id, vnid=None, value=None):
        params = {}
        if vnid != None:
            params['vnid'] = vnid
        return self.delete('v1/config/vrouter/{vrf_id}/routes'.format(vrf_id=vrf_id), value, params=params)

    def get_config_vrouter_vrf_id_routes(self, vrf_id, vnid=None, ip_prefix=None):
        params = {}
        if vnid != None:
            params['vnid'] = vnid
        if ip_prefix != None:
            params['ip_prefix'] = ip_prefix
        return self.get('v1/config/vrouter/{vrf_id}/routes'.format(vrf_id=vrf_id), params=params)

    def get_config_vrf_vrf_id_routes(self, vrf_id, ip_prefix=None):
        params = {}
        if ip_prefix != None:
            params['ip_prefix'] = ip_prefix
        return self.get('v1/config/vrf/{vrf_id}/routes'.format(vrf_id=vrf_id), params=params)

    # In memory DB restart
    def post_config_restart_in_mem_db(self):
        return self.post('v1/config/restartdb')

    # Operations
    # Ping
    def post_ping(self, value):
        return self.post('v1/operations/ping', value)

    # Helper functions
    def post_generic_vxlan_tunnel(self):
        rv = self.post_config_tunnel_decap_tunnel_type('vxlan', {
            'ip_addr': '34.53.1.0'
        })
        assert rv.status_code == 204

    def post_generic_vrouter_and_deps(self):
        self.post_generic_vxlan_tunnel()
        rv = self.post_config_vrouter_vrf_id("vnet-guid-1", {
            'vnid': 1001
        })
        assert rv.status_code == 204  

    def post_generic_vrouter_and_deps_duplicate(self):
        self.post_generic_vxlan_tunnel()
        rv = self.post_config_vrouter_vrf_id("vnet-guid-1", {
            'vnid': 1001
        })
        assert rv.status_code == 204
        rv = self.post_config_vrouter_vrf_id("vnet-guid-10", {
            'vnid': 1001
        })
        assert rv.status_code == 409 
        rv = self.post_config_vrouter_vrf_id("vnet-guid-1", {
            'vnid': 1001
        })
        assert rv.status_code == 409
        rv = self.post_config_vrouter_vrf_id("vnet-guid-1", {
            'vnid': 2001
        })
        assert rv.status_code == 409

    def post_generic_vlan_and_deps(self):
        self.post_generic_vrouter_and_deps()
        rv = self.post_config_vlan(2, {'vnet_id' : 'vnet-guid-1', 'ip_prefix' : '10.1.1.0/24'})
        assert rv.status_code == 204

    def check_routes_exist_in_tun_tb(self, vnet_num_mapped, routes_arr):
       for route in routes_arr:
           route_table = self.db.hgetall(ROUTE_TUN_TB + ':' + VNET_NAME_PREF +str(vnet_num_mapped)+':'+route['ip_prefix'])
           assert route_table == {
                            b'endpoint' : route['nexthop'],
                            b'mac_address' : route['mac_address'],
                            b'vni' : str(route['vnid'])
                          }

    def check_routes_dont_exist_in_tun_tb(self, vnet_num_mapped, routes_arr):
       for route in routes_arr:
           route_table = self.db.hgetall(ROUTE_TUN_TB + ':' + VNET_NAME_PREF +str(vnet_num_mapped)+':'+route['ip_prefix'])
           assert route_table == {}

    def check_routes_exist_in_loc_route_tb(self, vnet_num_mapped, routes_arr):
       for route in routes_arr:
           route_table = self.db.hgetall(LOCAL_ROUTE_TB + ':' + VNET_NAME_PREF +str(vnet_num_mapped)+':'+route['ip_prefix'])
           assert route_table == {
                            b'nexthop' : route['nexthop'],
                            b'ifname' : route['ifname']
                          }

    def check_routes_dont_exist_in_loc_route_tb(self, vnet_num_mapped, routes_arr):
       for route in routes_arr:
           route_table = self.db.hgetall(LOCAL_ROUTE_TB + ':' + VNET_NAME_PREF +str(vnet_num_mapped)+':'+route['ip_prefix'])
           assert route_table == {}
