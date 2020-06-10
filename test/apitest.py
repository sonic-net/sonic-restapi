#!/usr/bin/env python3

import datetime
import json
import requests
import time
import unittest
import logging
import redis
import json

TEST_HOST = 'http://localhost:8090/'

logging.basicConfig(filename='test.log', filemode='w', level=logging.INFO)
l = logging.getLogger('rest_api_test')

# DB Names
VXLAN_TUNNEL_TB   = "VXLAN_TUNNEL"
VNET_TB           = "VNET"
VLAN_TB           = "VLAN"
VLAN_INTF_TB      = "VLAN_INTERFACE"
VLAN_MEMB_TB      = "VLAN_MEMBER"
VLAN_NEIGH_TB     = "NEIGH"
ROUTE_TUN_TB      = "_VNET_ROUTE_TUNNEL_TABLE"
LOCAL_ROUTE_TB    = "_VNET_ROUTE_TABLE"
CFG_ROUTE_TUN_TB  = "VNET_ROUTE_TUNNEL"
CFG_LOCAL_ROUTE_TB = "VNET_ROUTE"

# DB Helper constants
VNET_NAME_PREF    = "Vnet"
VLAN_NAME_PREF    = "Vlan"

RESRC_EXISTS = 0
DEP_MISSING = 1
DELETE_DEP = 2

class rest_api_client(unittest.TestCase):
    maxDiff = None

    def post(self, url, body = []):
        if body == None:
            data = None
        else:
            data = json.dumps(body)

        l.info("Request POST: %s" % url)
        l.info("JSON Body: %s" % data)
        r = requests.post(TEST_HOST + url, data=data, headers={'Content-Type': 'application/json'})
        l.info('Response Code: %s' % r.status_code)
        l.info('Response Body: %s' % r.text)
        return r

    def patch(self, url, body = []):
        if body == None:
            data = None
        else:
            data = json.dumps(body)

        l.info("Request PATCH: %s" % url)
        l.info("JSON Body: %s" % data)
        r = requests.patch(TEST_HOST + url, data=data, headers={'Content-Type': 'application/json'})
        l.info('Response Code: %s' % r.status_code)
        l.info('Response Body: %s' % r.text)
        return r

    def get(self, url, body = [], params = {}):
        if body == None:
            data = None
        else:
            data = json.dumps(body)

        l.info("Request GET: %s" % url)
        l.info("JSON Body: %s" % data)
        r = requests.get(TEST_HOST + url, data=data, params=params, headers={'Content-Type': 'application/json'})
        l.info('Response Code: %s' % r.status_code)
        l.info('Response Body: %s' % r.text)
        return r

    def delete(self, url, body = [], params = {}):
        if body == None:
            data = None
        else:
            data = json.dumps(body)

        l.info("Request DELETE: %s" % url)
        l.info("JSON Body: %s" % data)
        r = requests.delete(TEST_HOST + url, data=data, params=params, headers={'Content-Type': 'application/json'})
        l.info('Response Code: %s' % r.status_code)
        l.info('Response Body: %s' % r.text)
        return r

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
        self.assertEqual(rv.status_code, 204)

    def post_generic_vrouter_and_deps(self):
        self.post_generic_vxlan_tunnel()
        rv = self.post_config_vrouter_vrf_id("vnet-guid-1", {
            'vnid': 1001
        })
        self.assertEqual(rv.status_code, 204)   

    def post_generic_vlan_and_deps(self):
        self.post_generic_vrouter_and_deps()
        rv = self.post_config_vlan(2, {'vnet_id' : 'vnet-guid-1', 'ip_prefix' : '10.1.1.0/24'})
        self.assertEqual(rv.status_code, 204)

    def check_routes_exist_in_db(self, vnet_num_mapped, routes_arr):
       for route in routes_arr:
           route_table = self.db.hgetall(ROUTE_TUN_TB + ':' + VNET_NAME_PREF +str(vnet_num_mapped)+':'+route['ip_prefix'])
           self.assertEqual(route_table, {
                            b'endpoint' : route['nexthop'],
                            b'mac_address' : route['mac_address'],
                            b'vni' : str(route['vnid'])
                          })

    def check_routes_dont_exist_in_db(self, vnet_num_mapped, routes_arr):
       for route in routes_arr:
           route_table = self.db.hgetall(ROUTE_TUN_TB + ':' + VNET_NAME_PREF +str(vnet_num_mapped)+':'+route['ip_prefix'])
           self.assertEqual(route_table, {})

    # Test setup
    def setUp(self):
        l.info('============================================================')
        l.info("Running: {0}".format(self._testMethodName))
        l.info('------------------------------------------------------------')

        # Clear DBs - reach known state
        self.db = redis.StrictRedis('localhost', 6379, 0)
        self.db.flushdb()

        self.cache = redis.StrictRedis('localhost', 6379, 7)
        self.cache.flushdb()

        self.configdb = redis.StrictRedis('localhost', 6379, 4)
        self.configdb.flushdb()

        # Sanity check
        keys = self.db.keys()
        self.assertEqual(keys, [])

        keys = self.cache.keys()
        self.assertEqual(keys, [])

        keys = self.configdb.keys()
        self.assertEqual(keys, [])

        self.post_config_restart_in_mem_db()


    @classmethod
    def setUpClass(cls):
        l.info('============================================================')
        l.info("Starting: {0} - {1}".format(cls.__name__, cls.__doc__))
        l.info('------------------------------------------------------------')


class ra_client_positive_tests(rest_api_client):
    """Normal behaviour tests"""
# Helper func
    def check_vrouter_exists(self, vnet_id, vnid):
        r = self.get_config_vrouter_vrf_id(vnet_id)
        self.assertEqual(r.status_code, 200)

        j = json.loads(r.text)
        self.assertEqual(j, {
            'vnet_id': vnet_id,
            'attr': {
                'vnid': vnid
            }
        })

    def helper_get_config_tunnel_decap_tunnel_type(self):
        self.post_generic_vxlan_tunnel()
        r = self.get_config_tunnel_decap_tunnel_type('vxlan')
        self.assertEqual(r.status_code, 200)

        j = json.loads(r.text)
        self.assertEqual(j, {
            'tunnel_type': 'vxlan',
            'attr': {
                'ip_addr': '34.53.1.0'
            }
        })

# Decap
    def test_post_config_tunnel_decap_tunnel_type(self):
        r = self.post_config_tunnel_decap_tunnel_type('vxlan', {
            'ip_addr': '34.53.1.0'
        })
        self.assertEqual(r.status_code, 204)

       # After 1st time config of decap, post is always no-op
        r = self.post_config_tunnel_decap_tunnel_type('vxlan', {
            'ip_addr': '74.32.6.0'
        })
        self.assertEqual(r.status_code, 204)

        tunnel_table = self.configdb.hgetall(VXLAN_TUNNEL_TB + '|default_vxlan_tunnel')
        self.assertEqual(tunnel_table, {b'src_ip': b'34.53.1.0'})
        l.info("Tunnel table is %s", tunnel_table)
        self.helper_get_config_tunnel_decap_tunnel_type()

    def test_delete_config_tunnel_decap_tunnel_type(self):
        self.post_generic_vxlan_tunnel()
        r = self.delete_config_tunnel_decap_tunnel_type('vxlan')
        self.assertEqual(r.status_code, 204)
        # The delete is a no-op and should return 204, moreover the tunnel should not be deleted 
        tunnel_table = self.configdb.hgetall(VXLAN_TUNNEL_TB + '|default_vxlan_tunnel')
        self.assertEqual(tunnel_table, {b'src_ip': b'34.53.1.0'})
        self.helper_get_config_tunnel_decap_tunnel_type()


# Encap
    def test_post_encap(self):
        r = self.post_config_tunnel_encap_vxlan_vnid(101, None)
        self.assertEqual(r.status_code, 204)
        keys = self.configdb.keys()
        self.assertEqual(keys, [])

    def test_get_encap(self):
        r = self.get_config_tunnel_encap_vxlan_vnid(101)
        self.assertEqual(r.status_code, 204)

    def test_delete_encap(self):
        r = self.delete_config_tunnel_encap_vxlan_vnid(101)
        self.assertEqual(r.status_code, 204)


# Vrouter
    def test_post_vrouter(self):
        self.post_generic_vxlan_tunnel()
        r = self.post_config_vrouter_vrf_id("vnet-guid-1", {
            'vnid': 1001
        })
        self.assertEqual(r.status_code, 204)

        vrouter_table = self.configdb.hgetall(VNET_TB + '|' + VNET_NAME_PREF + '1')
        self.assertEqual(vrouter_table, {
							b'vxlan_tunnel': b'default_vxlan_tunnel',
							b'vni': b'1001',
							b'guid': b'vnet-guid-1'
							})

    def  test_get_vrouter(self):
        self.post_generic_vrouter_and_deps()
        self.check_vrouter_exists("vnet-guid-1",1001)

    def test_delete_vrouter(self):
        self.post_generic_vrouter_and_deps()
        r = self.delete_config_vrouter_vrf_id("vnet-guid-1")
        self.assertEqual(r.status_code, 204)
        vrouter_table = self.configdb.hgetall(VNET_TB + '|' + VNET_NAME_PREF + '1')
        self.assertEqual(vrouter_table, {})

    def test_guid_persistence(self):
        self.post_generic_vrouter_and_deps()
        r = self.post_config_vrouter_vrf_id("vnet-guid-2", { 'vnid': 1002 })
        self.assertEqual(r.status_code, 204)
        r = self.post_config_vrouter_vrf_id("vnet-guid-3", { 'vnid': 1003 })
        self.assertEqual(r.status_code, 204)

        self.post_config_restart_in_mem_db()
        
        self.check_vrouter_exists("vnet-guid-1",1001)
        self.check_vrouter_exists("vnet-guid-2",1002)
        self.check_vrouter_exists("vnet-guid-3",1003)

    def test_vnet_name_mapping_logic(self):
        self.post_generic_vxlan_tunnel()
        for i in range (1,4):
             r = self.post_config_vrouter_vrf_id("vnet-guid-"+str(i), {'vnid': 1000+i})
             self.assertEqual(r.status_code, 204)
             self.check_vrouter_exists("vnet-guid-"+str(i), 1000+i)
             vrouter_table = self.configdb.hgetall(VNET_TB + '|' + VNET_NAME_PREF +str(i))
             self.assertEqual(vrouter_table, {
                     b'vxlan_tunnel': b'default_vxlan_tunnel',
                     b'vni': b'100'+str(i),
                     b'guid': b'vnet-guid-'+str(i)
                     })

        for i in range (1,4):
             r = self.delete_config_vrouter_vrf_id("vnet-guid-"+str(i))
             self.assertEqual(r.status_code, 204)
             r = self.post_config_vrouter_vrf_id("vnet-guid-"+str(i+3), {'vnid': 1003+i})
             self.assertEqual(r.status_code, 204)
             self.check_vrouter_exists("vnet-guid-"+str(i+3), 1003+i)
             vrouter_table = self.configdb.hgetall(VNET_TB + '|' + VNET_NAME_PREF +str(i))
             self.assertEqual(vrouter_table, {
                     b'vxlan_tunnel': b'default_vxlan_tunnel',
                     b'vni': b'100'+str(i+3),
                     b'guid': b'vnet-guid-'+str(i+3)
                     })

             r = self.post_config_vrouter_vrf_id("vnet-guid-"+str(i+6), {'vnid': 1006+i})
             self.assertEqual(r.status_code, 204)
             self.check_vrouter_exists("vnet-guid-"+str(i+6), 1006+i)
             vrouter_table = self.configdb.hgetall(VNET_TB + '|' + VNET_NAME_PREF +str(i+3))
             self.assertEqual(vrouter_table, {
                     b'vxlan_tunnel': b'default_vxlan_tunnel',
                     b'vni': b'100'+str(i+6),
                     b'guid': b'vnet-guid-'+str(i+6)
                     })
             

# Vlan
    def test_vlan_wo_ippref_vnetid_all_verbs(self):
        # post
        r = self.post_config_vlan(2, {})
        self.assertEqual(r.status_code, 204)
        
        # get
        r = self.get_config_vlan(2)
        self.assertEqual(r.status_code, 200)
        j = json.loads(r.text)
        self.assertEqual(j, {
            'vlan_id': 2,
            'attr': {}
        })
        vlan_table = self.configdb.hgetall(VLAN_TB + '|' + VLAN_NAME_PREF + '2')
        self.assertEqual(vlan_table, {b'vlanid': b'2'})
        vlan_intf_table = self.configdb.hgetall(VLAN_INTF_TB + '|' + VLAN_NAME_PREF + '2')
        self.assertEqual(vlan_intf_table, {})

        # delete
        r = self.delete_config_vlan(2)
        self.assertEqual(r.status_code, 204)
        vlan_table = self.configdb.hgetall(VLAN_TB + '|' + VLAN_NAME_PREF + '2')
        self.assertEqual(vlan_table, {})

    def test_vlan_with_vnetid_all_verbs(self):
        # post
        self.post_generic_vrouter_and_deps()
        r = self.post_config_vlan(2, {'vnet_id' : 'vnet-guid-1'})
        self.assertEqual(r.status_code, 204)
        
        # get
        r = self.get_config_vlan(2)
        self.assertEqual(r.status_code, 200)
        j = json.loads(r.text)
        self.assertEqual(j, {
            'vlan_id': 2,
            'attr': {'vnet_id':'vnet-guid-1'}
        })
        vlan_table = self.configdb.hgetall(VLAN_TB + '|' + VLAN_NAME_PREF + '2')
        self.assertEqual(vlan_table, {b'vlanid': b'2'})
        vlan_intf_table = self.configdb.hgetall(VLAN_INTF_TB + '|' + VLAN_NAME_PREF + '2')
        self.assertEqual(vlan_intf_table, {b'proxy_arp': b'enabled', b'vnet_name': VNET_NAME_PREF + '1'})

        # delete
        r = self.delete_config_vlan(2)
        self.assertEqual(r.status_code, 204)
        vlan_table = self.configdb.hgetall(VLAN_TB + '|' + VLAN_NAME_PREF + '2')
        self.assertEqual(vlan_table, {})
        vlan_intf_table = self.configdb.hgetall(VLAN_INTF_TB + '|' + VLAN_NAME_PREF + '2')
        self.assertEqual(vlan_intf_table, {})

    def test_vlan_with_ippref_all_verbs(self):
        # post
        self.post_generic_vrouter_and_deps()
        r = self.post_config_vlan(2, {'ip_prefix':'10.0.1.1/24'})
        self.assertEqual(r.status_code, 204)

        # get
        r = self.get_config_vlan(2)
        self.assertEqual(r.status_code, 200)
        j = json.loads(r.text)
        self.assertEqual(j, {
            'vlan_id': 2,
            'attr': {'ip_prefix':'10.0.1.1/24'}
        })
        vlan_table = self.configdb.hgetall(VLAN_TB + '|' + VLAN_NAME_PREF + '2')
        self.assertEqual(vlan_table, {b'vlanid': b'2'})
        vlan_intf_table = self.configdb.hgetall(VLAN_INTF_TB + '|' + VLAN_NAME_PREF + '2|10.0.1.1/24')
        self.assertEqual(vlan_intf_table, {b'':b''})
        vlan_intf_table = self.configdb.hgetall(VLAN_INTF_TB + '|' + VLAN_NAME_PREF + '2')
        self.assertEqual(vlan_intf_table, {})

        # delete
        r = self.delete_config_vlan(2)
        self.assertEqual(r.status_code, 204)
        vlan_table = self.configdb.hgetall(VLAN_TB + '|' + VLAN_NAME_PREF + '2')
        self.assertEqual(vlan_table, {})
        vlan_intf_table = self.configdb.hgetall(VLAN_INTF_TB + '|' + VLAN_NAME_PREF + '2|10.0.1.1/24')
        self.assertEqual(vlan_intf_table, {})

    def test_vlan_all_args_all_verbs(self):
        # post
        self.post_generic_vrouter_and_deps()
        r = self.post_config_vlan(2, {'vnet_id' : 'vnet-guid-1', 'ip_prefix':'10.0.1.1/24'})
        self.assertEqual(r.status_code, 204)

        # get
        r = self.get_config_vlan(2)
        self.assertEqual(r.status_code, 200)
        j = json.loads(r.text)
        self.assertEqual(j, {
            'vlan_id': 2,
            'attr': {'vnet_id' : 'vnet-guid-1', 'ip_prefix':'10.0.1.1/24'}
        })
        vlan_table = self.configdb.hgetall(VLAN_TB + '|' + VLAN_NAME_PREF + '2')
        self.assertEqual(vlan_table, {b'vlanid': b'2'})
        vlan_intf_table = self.configdb.hgetall(VLAN_INTF_TB + '|' + VLAN_NAME_PREF + '2|10.0.1.1/24')
        self.assertEqual(vlan_intf_table, {b'':b''})
        vlan_intf_table = self.configdb.hgetall(VLAN_INTF_TB + '|' + VLAN_NAME_PREF + '2')
        self.assertEqual(vlan_intf_table, {b'proxy_arp': b'enabled', b'vnet_name': VNET_NAME_PREF+'1'})

        # delete
        r = self.delete_config_vlan(2)
        self.assertEqual(r.status_code, 204)
        vlan_table = self.configdb.hgetall(VLAN_TB + '|' + VLAN_NAME_PREF + '2')
        self.assertEqual(vlan_table, {})
        vlan_intf_table = self.configdb.hgetall(VLAN_INTF_TB + '|' + VLAN_NAME_PREF + '2')
        self.assertEqual(vlan_intf_table, {})
        vlan_intf_table = self.configdb.hgetall(VLAN_INTF_TB + '|' + VLAN_NAME_PREF + '2|10.0.1.1/24')
        self.assertEqual(vlan_intf_table, {})

    def test_get_vlans_per_vnetid_1digitvlans(self):
        # create vxlan tunnel
        self.post_config_tunnel_decap_tunnel_type('vxlan', {
        'ip_addr': '6.6.6.6'
        })
        # create vnet_id/vrf
        self.post_config_vrouter_vrf_id('vnet-guid-1', {'vnid': 1001})
        self.post_config_vrouter_vrf_id('vnet-guid-2', {'vnid': 2001})
        #create vlan interfaces
        self.post_config_vlan(3, {'vnet_id' : 'vnet-guid-1', 'ip_prefix':'10.0.4.1/24'})
        self.post_config_vlan(4, {'vnet_id' : 'vnet-guid-1', 'ip_prefix':'10.0.3.1/24'})

        self.post_config_vlan(5, {'vnet_id' : 'vnet-guid-2', 'ip_prefix':'10.2.4.1/24'})
        self.post_config_vlan(6, {'vnet_id' : 'vnet-guid-2', 'ip_prefix':'10.2.3.1/24'})
        # get vlans for vnet-guid-1
        r_vnet1 = self.get_config_interface_vlans('vnet-guid-1')
        r_vnet2 = self.get_config_interface_vlans('vnet-guid-2')
        j_vnet1 = json.loads(r_vnet1.text)
        j_vnet2 = json.loads(r_vnet2.text)
        k_vnet1 = {"vnet_id":"vnet-guid-1","attr":[{"vlan_id":3,"ip_prefix":"10.0.4.1/24"},{"vlan_id":4,"ip_prefix":"10.0.3.1/24"}]}
        k_vnet2 = {"vnet_id":"vnet-guid-2","attr":[{"vlan_id":5,"ip_prefix":"10.2.4.1/24"},{"vlan_id":6,"ip_prefix":"10.2.3.1/24"}]}
        for key,value in j_vnet1.iteritems():
            if type(value)!=list:
                #print("not type list",value)
                self.assertEqual(k_vnet1[key],j_vnet1[key])
            else:
                #print("is type list",value)
                self.assertItemsEqual(value,k_vnet1.values()[0])
        for key,value in j_vnet2.iteritems():
            if type(value)!=list:
                #print("not type list",value)
                self.assertEqual(k_vnet2[key],j_vnet2[key])
            else:
                #print("is type list",value)
                self.assertItemsEqual(value,k_vnet2.values()[0])

    def test_get_vlans_per_vnetid_4digitvlans(self):
        # create vxlan tunnel
        self.post_config_tunnel_decap_tunnel_type('vxlan', {
        'ip_addr': '6.6.6.6'
        })
        # create vnet_id/vrf
        self.post_config_vrouter_vrf_id('vnet-guid-1', {'vnid': 1001})
        self.post_config_vrouter_vrf_id('vnet-guid-2', {'vnid': 2002})
        #create vlan interfaces
        self.post_config_vlan(1111, {'vnet_id' : 'vnet-guid-1', 'ip_prefix':'10.0.1.1/24'})
        self.post_config_vlan(2222, {'vnet_id' : 'vnet-guid-1', 'ip_prefix':'10.0.2.1/24'})
        self.post_config_vlan(3000, {'vnet_id' : 'vnet-guid-2'})
        self.post_config_vlan(4000, {'vnet_id' : 'vnet-guid-2', 'ip_prefix':'10.2.2.1/24'})

        # get vlans for vnet-guid-1
        r = self.get_config_interface_vlans('vnet-guid-1')
        j = json.loads(r.text)
        r2 = self.get_config_interface_vlans('vnet-guid-2')
        j2 = json.loads(r2.text)
        k = {"vnet_id":"vnet-guid-1","attr":[{"vlan_id":1111,"ip_prefix":"10.0.1.1/24"},{"vlan_id":2222,"ip_prefix":"10.0.2.1/24"}]}
        k2 = {"vnet_id":"vnet-guid-2","attr":[{"vlan_id":3000},{"vlan_id":4000,"ip_prefix":"10.2.2.1/24"}]}
        for key,value in j.iteritems():
            if type(value)!=list:
                #print("not type list",value)
                self.assertEqual(k[key],j[key])
            else:
                #print("is type list",value)
                self.assertItemsEqual(value,k.values()[0]) 
        for key,value in j2.iteritems():
            if type(value)!=list:
                #print("not type list",value)
                self.assertEqual(k2[key],j2[key])
            else:
                #print("is type list",value)
                self.assertItemsEqual(value,k2.values()[0])

# Vlan Get
    def test_get_all_vlans(self):
        # create vxlan tunnel
        self.post_config_tunnel_decap_tunnel_type('vxlan', {
        'ip_addr': '6.6.6.6'
        })
        # create vnet_id/vrf
        self.post_config_vrouter_vrf_id('vnet-guid-1', {'vnid': 1001})
        #create vlan interfaces
        self.post_config_vlan(3000, {'vnet_id' : 'vnet-guid-1', 'ip_prefix':'10.0.1.1/24'})
        self.post_config_vlan(3001, {'vnet_id' : 'vnet-guid-1'})

        # get all vlans
        r = self.get_config_vlans_all()
        j = json.loads(r.text)
        k = {"attr":[{"vlan_id":3000,"ip_prefix":"10.0.1.1/24","vnet_id":"vnet-guid-1"},{"vlan_id":3001,"vnet_id":"vnet-guid-1"}]}
        for key,value in j.iteritems():
            if type(value)!=list:
                self.assertEqual(k[key],j[key])
                return
            for item in k[key]:
                if item not in value:
                    assert False

# Vlan Member
    def test_vlan_member_tagged_untagged_interop(self):
        vlan0 = 2
        vlans = [3,4]
        members = ["Ethernet2", "Ethernet3", "Ethernet4"]
        self.post_generic_vrouter_and_deps()
        r = self.post_config_vlan(vlan0, {'vnet_id' : 'vnet-guid-1', 'ip_prefix':'10.0.1.1/24'})
        self.assertEqual(r.status_code, 204)
        for member in members:
           r = self.post_config_vlan_member(vlan0, member, {'tagging_mode' : 'untagged'})
           self.assertEqual(r.status_code, 204)
           r = self.get_config_vlan_member(vlan0,  member)
           self.assertEqual(r.status_code, 200)
           j = json.loads(r.text)
           self.assertEqual(j, {
                      'vlan_id': vlan0,
                      'if_name': member,
                      'attr': {'tagging_mode' : 'untagged'}
           })
        for vlan in vlans:
             r = self.post_config_vlan(vlan, {'vnet_id' : 'vnet-guid-1', 'ip_prefix':'10.0.1.1/24'})
             self.assertEqual(r.status_code, 204)
             # post
             for member in members:
                 r = self.post_config_vlan_member(vlan, member, {'tagging_mode' : 'tagged'})
                 self.assertEqual(r.status_code, 204)

                 # get
                 r = self.get_config_vlan_member(vlan,  member)
                 self.assertEqual(r.status_code, 200)
                 j = json.loads(r.text)
                 self.assertEqual(j, {
                      'vlan_id': vlan,
                      'if_name': member,
                      'attr': {'tagging_mode' : 'tagged'}
                 })

    def test_vlan_member_tagging_all_verbs(self):
        vlans = [2,3]
        members = ['Ethernet2', 'Ethernet3']
        self.post_generic_vrouter_and_deps()
        for vlan in vlans:
             r = self.post_config_vlan(vlan, {'vnet_id' : 'vnet-guid-1', 'ip_prefix':'10.0.1.1/24'})
             self.assertEqual(r.status_code, 204)
             # post
             for member in members:
                 r = self.post_config_vlan_member(vlan, member, {'tagging_mode' : 'tagged'})
                 self.assertEqual(r.status_code, 204)

                 # get
                 r = self.get_config_vlan_member(vlan,  member)
                 self.assertEqual(r.status_code, 200)
                 j = json.loads(r.text)
                 self.assertEqual(j, {
                      'vlan_id': vlan,
                      'if_name': member,
                      'attr': {'tagging_mode' : 'tagged'}
                 })
                 vlan_mem_table = self.configdb.hgetall(VLAN_MEMB_TB + '|' + VLAN_NAME_PREF +str(vlan)+'|'+member)
                 self.assertEqual(vlan_mem_table, {b'tagging_mode':b'tagged'})

                 # delete
                 r = self.delete_config_vlan_member(vlan, member)
                 self.assertEqual(r.status_code, 204)
                 vlan_mem_table = self.configdb.hgetall(VLAN_MEMB_TB + '|' + VLAN_NAME_PREF + str(vlan) + "|" +  member)
                 self.assertEqual(vlan_mem_table, {})

    def test_vlan_member_notagging_all_verbs(self):
        # post
        self.post_generic_vlan_and_deps()
        r = self.post_config_vlan_member(2, "Ethernet2", {})
        self.assertEqual(r.status_code, 204)

        # get
        r = self.get_config_vlan_member(2, "Ethernet2")
        self.assertEqual(r.status_code, 200)
        j = json.loads(r.text)
        self.assertEqual(j, {
            'vlan_id': 2,
            'if_name': 'Ethernet2',
            'attr': {'tagging_mode' : 'untagged'}
        })
        vlan_mem_table = self.configdb.hgetall(VLAN_MEMB_TB + '|' + VLAN_NAME_PREF + '2|Ethernet2')
        self.assertEqual(vlan_mem_table, {b'tagging_mode':b'untagged'})

    def test_get_members_per_vlan(self):
        # create vxlan tunnel
        self.post_config_tunnel_decap_tunnel_type('vxlan', {
        'ip_addr': '6.6.6.6'
        })
        # create vnet_id/vrf
        self.post_config_vrouter_vrf_id('vnet-guid-1', {'vnid': 1001})
        #create vlan interface 2
        self.post_config_vlan(2, {'vnet_id' : 'vnet-guid-1', 'ip_prefix':'10.0.1.1/24'})
        members = ["Ethernet2", "Ethernet3", "Ethernet4"]
        for member in members:
            self.post_config_vlan_member(2, member, {'tagging_mode' : 'untagged'})
        r = self.get_config_interface_vlan_members(2)
        j = json.loads(r.text)
        self.assertItemsEqual( j,
            {"vlan_id":2,"attr":[{"if_name":"Ethernet2","tagging_mode":"untagged"},{"if_name":"Ethernet3","tagging_mode":"untagged"},{"if_name":"Ethernet4","tagging_mode":"untagged"}]}
            )


# Vlan Neighbor
    def test_vlan_neighbor_all_verbs(self):
        # post
        self.post_generic_vlan_and_deps()
        r = self.post_config_vlan_neighbor(2, "10.10.10.10")
        self.assertEqual(r.status_code, 204)

        # get
        r = self.get_config_vlan_neighbor(2, "10.10.10.10")
        self.assertEqual(r.status_code, 200)
        j = json.loads(r.text)
        self.assertEqual(j, {
            'vlan_id': 2,
            'ip_addr': '10.10.10.10'
        })
        vlan_neigh_table = self.configdb.hgetall(VLAN_NEIGH_TB + '|' + VLAN_NAME_PREF + '2|10.10.10.10')
        self.assertEqual(vlan_neigh_table, {b'family':b'IPv4'})

        # delete
        r = self.delete_config_vlan_neighbor(2, "10.10.10.10")
        self.assertEqual(r.status_code, 204)
        vlan_neigh_table = self.configdb.hgetall(VLAN_NEIGH_TB + '|' + VLAN_NAME_PREF + '2|10.10.10.10')
        self.assertEqual(vlan_neigh_table, {})

    def test_get_neighbors_per_vlan(self):
        # create vxlan tunnel
        self.post_config_tunnel_decap_tunnel_type('vxlan', {
        'ip_addr': '6.6.6.6'
        })
        # create vnet_id/vrf
        self.post_config_vrouter_vrf_id('vnet-guid-1', {'vnid': 1001})
        #create vlan interface 2
        self.post_config_vlan(3, {'vnet_id' : 'vnet-guid-1', 'ip_prefix':'10.0.2.1/24'})
        self.post_config_vlan_neighbor(3, "10.10.20.10")
        self.post_config_vlan_neighbor(3, "10.10.30.10")

        # get vlans for vnet-guid-1
        r = self.get_config_interface_vlan_neighbors(3)
        j = json.loads(r.text)
        self.assertItemsEqual( j,
            {"vlan_id":3,"attr":[{"ip_addr":"10.10.20.10"},{"ip_addr":"10.10.30.10"}]}
            )

# Routes
    def test_patch_update_routes_with_optional_args(self):
        self.post_generic_vlan_and_deps()
        # No optional args
        route = {
                 'cmd':'add',
		 'ip_prefix':'10.2.1.0/24',
                 'nexthop':'192.168.2.1'
                }
        r = self.patch_config_vrouter_vrf_id_routes("vnet-guid-1", [route])
        self.assertEqual(r.status_code, 204)
        route_table = self.db.hgetall(ROUTE_TUN_TB + ':' + VNET_NAME_PREF +str(1)+':'+route['ip_prefix'])
        self.assertEqual(route_table, {b'endpoint' : route['nexthop']})
        del route['cmd']
        r = self.get_config_vrouter_vrf_id_routes("vnet-guid-1")
        self.assertEqual(r.status_code, 200)
        j = json.loads(r.text)
        self.assertItemsEqual(j, [route])

        # Vnid Optional arg
        route['vnid'] = 5000
        route['cmd'] = 'add'
        r = self.patch_config_vrouter_vrf_id_routes("vnet-guid-1", [route])
        self.assertEqual(r.status_code, 204)
        route_table = self.db.hgetall(ROUTE_TUN_TB + ':' + VNET_NAME_PREF +str(1)+':'+route['ip_prefix'])
        self.assertEqual(route_table, {b'endpoint' : route['nexthop'],
                                       b'vni' : str(route['vnid'])
                                      })
        del route['cmd']
        r = self.get_config_vrouter_vrf_id_routes("vnet-guid-1")
        self.assertEqual(r.status_code, 200)
        j = json.loads(r.text)
        self.assertItemsEqual(j, [route])       

        # Mac address Optional arg
        del route['vnid']
        route['mac_address'] = '00:08:aa:bb:cd:ef'
        route['cmd'] = 'add'
        r = self.patch_config_vrouter_vrf_id_routes("vnet-guid-1", [route])
        self.assertEqual(r.status_code, 204)
        route_table = self.db.hgetall(ROUTE_TUN_TB + ':' + VNET_NAME_PREF +str(1)+':'+route['ip_prefix'])
        self.assertEqual(route_table, {b'endpoint' : route['nexthop'],
                                       b'mac_address' : route['mac_address']
                                      })
        del route['cmd']
        r = self.get_config_vrouter_vrf_id_routes("vnet-guid-1")
        self.assertEqual(r.status_code, 200)
        j = json.loads(r.text)
        self.assertItemsEqual(j, [route])


    def test_patch_routes_drop_bm_routes(self):
        cidr = [24,30,32]
        self.post_generic_vrouter_and_deps()
        rv = self.post_config_vlan(2, {'vnet_id' : 'vnet-guid-1', 'ip_prefix' : '10.1.1.0/24'})
        self.assertEqual(rv.status_code, 204)
        r = self.post_config_vlan(3, {'vnet_id' : 'vnet-guid-1', 'ip_prefix':'10.1.5.0/24'})
        self.assertEqual(r.status_code, 204)
        routes = []
        for i in range (1,7):
             for ci in cidr:
		routes.append({'cmd':'add',
                            'ip_prefix':'10.1.'+str(i)+'.1/'+str(ci),
                            'nexthop':'192.168.2.'+str(i),
                            'vnid': 1 + i%5,
                            'mac_address':'00:08:aa:bb:cd:'+hex(15+i)[2:]})

        r = self.patch_config_vrouter_vrf_id_routes("vnet-guid-1", routes)
        self.assertEqual(r.status_code, 204)

        routes_bm = []
        routes_not_bm = []
        for route in routes:
             del route['cmd']
             if route['ip_prefix'] == '10.1.1.1/32' or  route['ip_prefix'] == '10.1.5.1/32':
                  routes_bm.append(route)
             else:
                  routes_not_bm.append(route)
        self.check_routes_exist_in_db(1, routes_not_bm) 
        self.check_routes_dont_exist_in_db(1, routes_bm) 

        r = self.get_config_vrouter_vrf_id_routes("vnet-guid-1")
        self.assertEqual(r.status_code, 200)
        j = json.loads(r.text)
        self.assertItemsEqual(j, routes_not_bm)

        for route in routes_bm:
             route['cmd'] = 'delete'
        r = self.patch_config_vrouter_vrf_id_routes("vnet-guid-1", routes_bm)
        self.assertEqual(r.status_code, 204)
        

       
    def test_routes_all_verbs(self):
        self.post_generic_vlan_and_deps()
        routes = []
        rv = self.post_config_vrouter_vrf_id("vnet-guid-2", {
            'vnid': 1002
        })
        self.assertEqual(rv.status_code, 204)
        for i in range (1,100):
             routes.append({'cmd':'add', 
                            'ip_prefix':'10.2.'+str(i)+'.0/24', 
                            'nexthop':'192.168.2.'+str(i), 
                            'vnid': 1 + i%5, 
                            'mac_address':'00:08:aa:bb:cd:'+hex(15+i)[2:]})
       
        # Patch add
        r = self.patch_config_vrouter_vrf_id_routes("vnet-guid-1", routes)
        self.assertEqual(r.status_code, 204)
        self.check_routes_exist_in_db(1, routes)
        r = self.patch_config_vrouter_vrf_id_routes("vnet-guid-2", routes)
        self.assertEqual(r.status_code, 204)
        self.check_routes_exist_in_db(2, routes)
        # Patch delete

        # Get all
        for route in routes:
             del route['cmd']
        r = self.get_config_vrouter_vrf_id_routes("vnet-guid-1")
        self.assertEqual(r.status_code, 200)
        j = json.loads(r.text)
        self.assertItemsEqual(j, routes)

        # Get filtered by vnid
        routes_vnid = []
        routes_not_vnid = []
        route_pref = {}
        i = 0
        for route in routes:
            if i == 70:
                 route_pref = route
            if route['vnid'] == 5:
                 routes_vnid.append(route)
            else:
                 routes_not_vnid.append(route)
            i += 1
        r = self.get_config_vrouter_vrf_id_routes("vnet-guid-1", vnid=5)
        self.assertEqual(r.status_code, 200)
        j = json.loads(r.text)
        self.assertItemsEqual(j, routes_vnid)

        # Get filtered by ip_prefix
        r = self.get_config_vrouter_vrf_id_routes("vnet-guid-1", ip_prefix=route_pref['ip_prefix'])
        self.assertEqual(r.status_code, 200)
        j = json.loads(r.text)
        self.assertItemsEqual(j, [route_pref])
 
        # Get filtered by both ip_prefix and vnid
        r = self.get_config_vrouter_vrf_id_routes("vnet-guid-1", vnid=2, ip_prefix=route_pref['ip_prefix'])
        self.assertEqual(r.status_code, 200)
        j = json.loads(r.text)
        self.assertItemsEqual(j, [route_pref])

        # Delete filtered by vnid
        r = self.delete_config_vrouter_vrf_id_routes("vnet-guid-1", vnid=5)
        self.assertEqual(r.status_code, 204)
        self.check_routes_exist_in_db(1, routes_not_vnid)
        self.check_routes_dont_exist_in_db(1, routes_vnid)
        r = self.get_config_vrouter_vrf_id_routes("vnet-guid-1", vnid=5)
        j = json.loads(r.text)
        self.assertEqual(j, [])

        # Patch combo add and delete
        for route in routes:
              if route['vnid'] == 5:
                    route['cmd'] = 'add'
              else:
                    route['cmd'] = 'delete'
        r = self.patch_config_vrouter_vrf_id_routes("vnet-guid-1", routes)
        self.assertEqual(r.status_code, 204)
        self.check_routes_exist_in_db(1, routes_vnid)
        self.check_routes_dont_exist_in_db(1, routes_not_vnid)
        for route in routes:
             del route['cmd']
        r = self.get_config_vrouter_vrf_id_routes("vnet-guid-1")
        self.assertEqual(r.status_code, 200)
        j = json.loads(r.text)
        self.assertItemsEqual(j, routes_vnid)
        r = self.get_config_vrouter_vrf_id_routes("vnet-guid-1", vnid=4)
        self.assertEqual(r.status_code, 200)
        j = json.loads(r.text)
        self.assertEqual(j, [])

        # Delete all routes
        r = self.delete_config_vrouter_vrf_id_routes("vnet-guid-1")
        self.assertEqual(r.status_code, 204)
        self.check_routes_dont_exist_in_db(1, routes)
        r = self.get_config_vrouter_vrf_id_routes("vnet-guid-1")
        j = json.loads(r.text)
        self.assertEqual(j, [])

        # Test that routes in other Vnet are untouched
        self.check_routes_exist_in_db(2, routes)
        r = self.get_config_vrouter_vrf_id_routes("vnet-guid-2")
        self.assertEqual(r.status_code, 200)
        j = json.loads(r.text)
        self.assertItemsEqual(j, routes)
        
    def test_local_subnet_route_addition(self):
        self.post_generic_vlan_and_deps()
        local_route_table = self.db.hgetall(LOCAL_ROUTE_TB + ':' + VNET_NAME_PREF +str(1)+':10.1.1.0/24')
        self.assertEqual(local_route_table, {b'ifname' : VLAN_NAME_PREF + '2'})
        r = self.delete_config_vlan(2)
        self.assertEqual(r.status_code, 204)
        local_route_table = self.db.hgetall(LOCAL_ROUTE_TB + ':' + VNET_NAME_PREF +str(1)+':10.1.1.0/24')
        self.assertEqual(local_route_table, {})

    # Operations
    # PingVRF
    def test_post_ping(self):
        vlan0 = 2
        self.post_generic_vrouter_and_deps()
        # Ping loss but response 200
        r = self.post_ping({'vnet_id' : 'vnet-guid-1', 'count' : '2', 'ip_addr' : '8.8.8.8'})
        self.assertEqual(r.status_code, 200)
        # Ping success and response 200
        r = self.post_ping({"count" : "2", "ip_addr" : "8.8.8.8"})
        self.assertEqual(r.status_code, 200)
        # Ping success and response 200
        r = self.post_ping({"ip_addr" : "8.8.8.8"})
        self.assertEqual(r.status_code, 200)
        
class ra_client_negative_tests(rest_api_client):
    """Invalid input tests"""
# Decap:
    def test_delete_config_tunnel_decap_tunnel_type_not_vxlan(self):
        r = self.delete_config_tunnel_decap_tunnel_type('not_vxlan')
        self.assertEqual(r.status_code, 400)

        j = json.loads(r.text)
        self.assertListEqual(['tunnel_type'], j['error']['fields'])

    def test_get_config_tunnel_decap_tunnel_not_created(self):
        r = self.get_config_tunnel_decap_tunnel_type('vxlan')
        self.assertEqual(r.status_code, 404)

        j = json.loads(r.text)
        self.assertListEqual(['tunnel_type'], j['error']['fields'])

# Vrouter: 
    def test_delete_vrouter_with_dependencies(self):
        # Init
        self.post_generic_vrouter_and_deps()
        r = self.post_config_vrouter_vrf_id("vnet-guid-2", { 'vnid': 1002 })
        self.assertEqual(r.status_code, 204)
        
        # Vlan Dependency
        rv = self.post_config_vlan(2, {'vnet_id' : 'vnet-guid-1', 'ip_prefix' : '10.1.1.0/24'})
        self.assertEqual(rv.status_code, 204)
        rv = self.post_config_vlan(3,  {'ip_prefix' : '10.1.1.0/24'})
        self.assertEqual(rv.status_code, 204)
        r = self.delete_config_vrouter_vrf_id("vnet-guid-1")
        self.assertEqual(r.status_code, 409)
        j = json.loads(r.text)
        self.assertEqual(DELETE_DEP, j['error']['sub-code'])
        rv = self.delete_config_vlan(2)
        self.assertEqual(rv.status_code, 204)
        r = self.delete_config_vrouter_vrf_id("vnet-guid-1")
        self.assertEqual(r.status_code, 204)

        # Routes Dependency
        self.post_generic_vrouter_and_deps()
        r = self.patch_config_vrouter_vrf_id_routes("vnet-guid-1", [{'cmd':'add', 'ip_prefix':'10.1.2.0/24', 'nexthop':'192.168.2.1'}])
        self.assertEqual(r.status_code, 204)
        r = self.delete_config_vrouter_vrf_id("vnet-guid-1")
        self.assertEqual(r.status_code, 409)
        j = json.loads(r.text)
        self.assertEqual(DELETE_DEP, j['error']['sub-code'])
        r = self.patch_config_vrouter_vrf_id_routes("vnet-guid-1", [{'cmd':'delete', 'ip_prefix':'10.1.2.0/24', 'nexthop':'192.168.2.1'}])
        self.assertEqual(r.status_code, 204)
        r = self.delete_config_vrouter_vrf_id("vnet-guid-1")
        self.assertEqual(r.status_code, 204)


    def test_vrouter_not_created_all_verbs(self):
        # Get
        r = self.get_config_vrouter_vrf_id("vnet-guid-1")
        self.assertEqual(r.status_code, 404)

        # Delete
        r = self.delete_config_vrouter_vrf_id("vnet-guid-1")
        self.assertEqual(r.status_code, 404)

        # Vrouter Routes
        r = self.patch_config_vrouter_vrf_id_routes("vnet-guid-1", [{'cmd':'add', 'ip_prefix':'10.1.2.0/24', 'nexthop':'192.168.2.1'}])
        self.assertEqual(r.status_code, 404)
        j = json.loads(r.text)
        self.assertListEqual(['vnet-guid-1'], j['error']['fields'])
        r = self.get_config_vrouter_vrf_id_routes("vnet-guid-1")
        self.assertEqual(r.status_code, 404)
        j = json.loads(r.text)
        self.assertListEqual(['vnet-guid-1'], j['error']['fields'])
        r = self.delete_config_vrouter_vrf_id_routes("vnet-guid-1")
        self.assertEqual(r.status_code, 404)
        j = json.loads(r.text)
        self.assertListEqual(['vnet-guid-1'], j['error']['fields'])

    def test_post_vrouter_without_vtep(self):
        r = self.post_config_vrouter_vrf_id("vnet-guid-1", {
            'vnid': 1001
        })
        self.assertEqual(r.status_code, 409)

        j = json.loads(r.text)
        self.assertListEqual(['tunnel'], j['error']['fields'])
        self.assertEqual(DEP_MISSING, j['error']['sub-code'])

    def test_post_vrouter_which_exists(self):
        self.post_generic_vrouter_and_deps()
        r = self.post_config_vrouter_vrf_id("vnet-guid-1", {
            'vnid': 1001
        })
        self.assertEqual(r.status_code, 409)

        j = json.loads(r.text)
        self.assertEqual(RESRC_EXISTS, j['error']['sub-code'])

    def test_post_vrouter_malformed_arg(self):
        self.post_generic_vrouter_and_deps()
        r = self.post_config_vrouter_vrf_id("vnet-guid-1", {
            'vnid': "this is malformed"
        })
        self.assertEqual(r.status_code, 400)

        j = json.loads(r.text)
        self.assertListEqual(['vnid'], j['error']['fields'])

# Vlan
    def test_post_vlan_which_exists(self):
        self.post_generic_vlan_and_deps()
        r = self.post_config_vlan(2, {})
        self.assertEqual(r.status_code, 409)

        j = json.loads(r.text)
        self.assertEqual(RESRC_EXISTS, j['error']['sub-code'])

    def test_vlan_not_created_all_verbs(self):
        # Get
        r = self.get_config_vlan(2)
        self.assertEqual(r.status_code, 404) 
        r = self.get_config_vlan_member(2, "ethernet2")
        self.assertEqual(r.status_code, 404)
        j = json.loads(r.text)
        self.assertListEqual(['vlan_id'], j['error']['fields'])
        r = self.get_config_vlan_neighbor(2, "10.10.10.10")
        self.assertEqual(r.status_code, 404)
        j = json.loads(r.text)
        self.assertListEqual(['vlan_id'], j['error']['fields'])

        # Delete
        r = self.delete_config_vlan(2)
        self.assertEqual(r.status_code, 404)
        r = self.delete_config_vlan_member(2, "ethernet2")
        self.assertEqual(r.status_code, 404)
        j = json.loads(r.text)
        self.assertListEqual(['vlan_id'], j['error']['fields'])
        r = self.delete_config_vlan_neighbor(2, "10.10.10.10")
        self.assertEqual(r.status_code, 404)
        j = json.loads(r.text)
        self.assertListEqual(['vlan_id'], j['error']['fields'])

        # Post
        r = self.post_config_vlan_member(2, "ethernet2", {})
        self.assertEqual(r.status_code, 404)
        j = json.loads(r.text)
        self.assertListEqual(['vlan_id'], j['error']['fields'])
        r = self.post_config_vlan_neighbor(2, "10.10.10.10")
        self.assertEqual(r.status_code, 404)
        j = json.loads(r.text)
        self.assertListEqual(['vlan_id'], j['error']['fields'])

    def test_vlan_out_of_range(self):
       vlan_ids = [1,4095]
       member = "Ethernet1"
       ip_addr = "10.10.1.1"
       for vlan_id in vlan_ids:
            r = self.post_config_vlan(vlan_id, {})
            self.assertEqual(r.status_code, 400)
            r = self.get_config_vlan(vlan_id)
            self.assertEqual(r.status_code, 400)
            r = self.delete_config_vlan(vlan_id)
            self.assertEqual(r.status_code, 400)
            
            r = self.post_config_vlan_member(vlan_id, member, {})
            self.assertEqual(r.status_code, 400)
            r = self.get_config_vlan_member(vlan_id, member)
            self.assertEqual(r.status_code, 400)
            r = self.delete_config_vlan_member(vlan_id, member)         
            self.assertEqual(r.status_code, 400)

            r = self.post_config_vlan_neighbor(vlan_id, {})
            self.assertEqual(r.status_code, 400)
            r = self.get_config_vlan_neighbor(vlan_id, ip_addr)
            self.assertEqual(r.status_code, 400)
            r = self.delete_config_vlan_neighbor(vlan_id, ip_addr)
            self.assertEqual(r.status_code, 400)

    def test_delete_vlan_with_dependencies(self):
        # Init generic config
        self.post_generic_vlan_and_deps()
        rv = self.post_config_vlan(3, {'vnet_id' : 'vnet-guid-1', 'ip_prefix' : '10.1.1.0/24'})
        self.assertEqual(rv.status_code, 204)
        rv = self.post_config_vlan_member(3, "Ethernet2", {})
        self.assertEqual(rv.status_code, 204)

        # Dependency Vlan Member
        rv = self.post_config_vlan_member(2, "Ethernet1", {})
        self.assertEqual(rv.status_code, 204)
        r = self.delete_config_vlan(2)
        self.assertEqual(r.status_code, 409)
        j = json.loads(r.text)
        self.assertEqual(DELETE_DEP, j['error']['sub-code'])
        rv = self.delete_config_vlan_member(2, "Ethernet1")
        self.assertEqual(rv.status_code, 204)
        r = self.delete_config_vlan(2)
        self.assertEqual(r.status_code, 204)

        # Dependency Vlan Neighbor
        rv = self.post_config_vlan(2, {'vnet_id' : 'vnet-guid-1', 'ip_prefix' : '10.1.1.0/24'})
        rv = self.post_config_vlan_neighbor(2, "10.10.10.10")
        self.assertEqual(rv.status_code, 204)
        r = self.delete_config_vlan(2)
        self.assertEqual(r.status_code, 409)
        j = json.loads(r.text)
        self.assertEqual(DELETE_DEP, j['error']['sub-code'])
        rv = self.delete_config_vlan_neighbor(2, "10.10.10.10")
        self.assertEqual(rv.status_code, 204)
        r = self.delete_config_vlan(2)
        self.assertEqual(r.status_code, 204)

    def test_get_vlans_per_vnetid_invalid_vlan(self):
        # create vxlan tunnel
        self.post_config_tunnel_decap_tunnel_type('vxlan', {
        'ip_addr': '6.6.6.6'
        })
        # create vnet_id/vrf
        self.post_config_vrouter_vrf_id('vnet-guid-1', {'vnid': 1001})
        #create invalid vlan interfaces
        self.post_config_vlan(5555, {'vnet_id' : 'vnet-guid-1', 'ip_prefix':'10.0.1.1/24'})
        self.post_config_vlan(4096, {'vnet_id' : 'vnet-guid-1', 'ip_prefix':'10.0.2.1/24'})

        # get vlans for vnet-guid-1
        r = self.get_config_interface_vlans('vnet-guid-1')
        j = json.loads(r.text)
        self.assertEqual(j,{u'attr': None, u'vnet_id': u'vnet-guid-1'})

    def test_get_vlans_per_vnetid_invalid_vnet(self):
        # create vxlan tunnel
        self.post_config_tunnel_decap_tunnel_type('vxlan', {
        'ip_addr': '6.6.6.6'
        })
        # create vnet_id/vrf
        self.post_config_vrouter_vrf_id('vnet-guid-1', {'vnid': 1001})
        #create vlan interfaces
        self.post_config_vlan(555, {'vnet_id' : 'vnet-guid-1', 'ip_prefix':'10.0.1.1/24'})
        self.post_config_vlan(409, {'vnet_id' : 'vnet-guid-1', 'ip_prefix':'10.0.2.1/24'})

        # get vlans for vnet-guid-1
        r = self.get_config_interface_vlans('')
        j = json.loads(r.text)
        self.assertEqual(r.status_code,404)

# Vlan Member
    def test_post_vlan_mem_which_exists_tagged(self):
        self.post_generic_vlan_and_deps()
        r = self.post_config_vlan(3, {'vnet_id' : 'vnet-guid-1', 'ip_prefix' : '10.1.1.0/24'})
        self.assertEqual(r.status_code, 204)
        attr = {'tagging_mode' : 'tagged'}
        r = self.post_config_vlan_member(2, "Ethernet1", attr)
        self.assertEqual(r.status_code, 204)

        r = self.post_config_vlan_member(2, "Ethernet1", attr)
        self.assertEqual(r.status_code, 409)
        j = json.loads(r.text)
        self.assertEqual(RESRC_EXISTS, j['error']['sub-code'])

        r = self.post_config_vlan_member(3, "Ethernet1", attr)
        self.assertEqual(r.status_code, 204)

        
    def test_post_vlan_mem_which_exists_untagged(self):
        self.post_generic_vlan_and_deps()
        r = self.post_config_vlan(3, {'vnet_id' : 'vnet-guid-1', 'ip_prefix' : '10.1.1.0/24'})
        self.assertEqual(r.status_code, 204)
        attr = {'tagging_mode' : 'untagged'}

        r = self.post_config_vlan_member(2, "Ethernet1", attr)
        self.assertEqual(r.status_code, 204)

        r = self.post_config_vlan_member(2, "Ethernet1", attr)
        self.assertEqual(r.status_code, 409)
        j = json.loads(r.text)
        self.assertEqual(RESRC_EXISTS, j['error']['sub-code'])

        r = self.post_config_vlan_member(3, "Ethernet1", attr)
        self.assertEqual(r.status_code, 409)
        j = json.loads(r.text)
        self.assertEqual(RESRC_EXISTS, j['error']['sub-code'])

        r = self.delete_config_vlan_member(2, "Ethernet1")
        self.assertEqual(r.status_code, 204)
        r = self.post_config_vlan_member(3, "Ethernet1", attr)
        self.assertEqual(r.status_code, 204)

        attr = {'tagging_mode' : 'tagged'}
        r = self.post_config_vlan_member(2, "Ethernet1", attr)
        self.assertEqual(r.status_code, 204)

    def test_get_vlan_member_not_created(self):
        self.post_generic_vlan_and_deps()
        r = self.get_config_vlan_member(2, "ethernet2")
        self.assertEqual(r.status_code, 404)
        j = json.loads(r.text)
        self.assertListEqual(['if_name'], j['error']['fields'])

    def test_delete_vlan_member_not_created(self):
        self.post_generic_vlan_and_deps()
        r = self.delete_config_vlan_member(2, "ethernet2")
        self.assertEqual(r.status_code, 404)
        j = json.loads(r.text)
        self.assertListEqual(['if_name'], j['error']['fields'])

# Vlan Neighbor
    def test_post_vlan_neighbor_which_exists(self):
        self.post_generic_vlan_and_deps()
        r = self.post_config_vlan_neighbor(2, "10.10.10.10")
        self.assertEqual(r.status_code, 204)

        r = self.post_config_vlan_neighbor(2, "10.10.10.10")
        self.assertEqual(r.status_code, 409)

        j = json.loads(r.text)
        self.assertEqual(RESRC_EXISTS, j['error']['sub-code'])       

    def test_get_vlan_neighbor_not_created(self):
        self.post_generic_vlan_and_deps()
        r = self.get_config_vlan_neighbor(2, "10.10.10.10")
        self.assertEqual(r.status_code, 404)
        j = json.loads(r.text)
        self.assertListEqual(['ip_addr'], j['error']['fields'])

    def test_delete_vlan_neighbor_not_created(self):
        self.post_generic_vlan_and_deps()
        r = self.delete_config_vlan_neighbor(2, "10.10.10.10")
        self.assertEqual(r.status_code, 404)
        j = json.loads(r.text)
        self.assertListEqual(['ip_addr'], j['error']['fields'])

    def test_vlan_neighbor_not_valid_ip(self):
        self.post_generic_vlan_and_deps()
        # post
        r = self.post_config_vlan_neighbor(2, "a.b.c.d")
        self.assertEqual(r.status_code, 400)
        j = json.loads(r.text)
        self.assertListEqual(['ip_addr'], j['error']['fields'])

        # get
        r = self.get_config_vlan_neighbor(2, "a.b.c.d")
        self.assertEqual(r.status_code, 400)
        j = json.loads(r.text)
        self.assertListEqual(['ip_addr'], j['error']['fields'])
 
        # delete
        r = self.delete_config_vlan_neighbor(2, "a.b.c.d")
        self.assertEqual(r.status_code, 400)
        j = json.loads(r.text)
        self.assertListEqual(['ip_addr'], j['error']['fields'])

# Routes
    def test_patch_delete_routes_not_created(self):
        self.post_generic_vlan_and_deps()
        routes = []
        for i in range (1,100):
             routes.append({'cmd':'delete',
                            'ip_prefix':'10.2.'+str(i)+'.0/24',
                            'nexthop':'192.168.2.'+str(i),
                            'vnid': 1 + i%5,
                            'mac_address':'00:08:aa:bb:cd:'+hex(15+i)[2:]})

        # Patch
        r = self.patch_config_vrouter_vrf_id_routes("vnet-guid-1", routes)
        self.assertEqual(r.status_code, 207)
        j = json.loads(r.text)
        for route in routes:
             route['error_code'] = 404
             route['error_msg'] = 'Not found'
        self.assertItemsEqual(routes, j['failed'])
        self.check_routes_dont_exist_in_db(1, routes)

    # Operations
    # PingVRF
    def test_post_ping_invalid(self):
        vlan0 = 2
        self.post_generic_vrouter_and_deps()
        # Invalid count scenario
        r = self.post_ping({"count" : "abc", "ip_addr" : "8.8.8.8"})
        self.assertEqual(r.status_code, 400)
        # Invalid ip_addr scenario
        r = self.post_ping({"ip_addr" : "8.8.8.888"})
        self.assertEqual(r.status_code, 400)
        # vnet_id not found 404 error
        r = self.post_ping({'vnet_id' : 'vnet-1', 'ip_addr' : '8.8.8.8'})
        self.assertEqual(r.status_code, 404)


suite = unittest.TestLoader().loadTestsFromTestCase(ra_client_positive_tests)
unittest.TextTestRunner(verbosity=2).run(suite)

suite = unittest.TestLoader().loadTestsFromTestCase(ra_client_negative_tests)
unittest.TextTestRunner(verbosity=2).run(suite)
