#!/usr/bin/env python3

import logging
import json

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


class TestRestApiPositive:
    """Normal behaviour tests"""
    # Helper func
    def check_vrouter_exists(self, restapi_client, vnet_id, vnid):
        r = restapi_client.get_config_vrouter_vrf_id(vnet_id)
        assert r.status_code == 200
        j = json.loads(r.text)
        assert j == {
            'vnet_id': vnet_id,
            'attr': {
                'vnid': vnid
            }
        }

    def helper_get_config_tunnel_decap_tunnel_type(self, restapi_client):
        restapi_client.post_generic_vxlan_tunnel()
        r = restapi_client.get_config_tunnel_decap_tunnel_type('vxlan')
        assert r.status_code == 200

        j = json.loads(r.text)
        assert j == {
            'tunnel_type': 'vxlan',
            'attr': {
                'ip_addr': '34.53.1.0'
            }
        }

    # Config reset status
    def test_config_status_reset_get(self, setup_restapi_client):
        _, _, _, restapi_client = setup_restapi_client
        r = restapi_client.get_config_reset_status()
        assert r.status_code == 200
        j = json.loads(r.text)
        assert j == {
            'reset_status': 'true'
        }

    def test_config_status_reset_post(self, setup_restapi_client):
        _, _, _, restapi_client = setup_restapi_client
        r = restapi_client.post_config_reset_status({'reset_status': 'false'})
        assert r.status_code == 200
        j = json.loads(r.text)
        assert j == {
            'reset_status': 'false'
        }
        r = restapi_client.post_config_reset_status({'reset_status': 'boolean'})
        assert r.status_code == 400

    # Decap
    def test_post_config_tunnel_decap_tunnel_type(self, setup_restapi_client):
        _, _, configdb, restapi_client = setup_restapi_client
        r = restapi_client.post_config_tunnel_decap_tunnel_type('vxlan', {
            'ip_addr': '34.53.1.0'
        })
        assert r.status_code == 204

       # After 1st time config of decap, post is always no-op
        r = restapi_client.post_config_tunnel_decap_tunnel_type('vxlan', {
            'ip_addr': '74.32.6.0'
        })
        assert r.status_code == 409

        tunnel_table = configdb.hgetall(VXLAN_TUNNEL_TB + '|default_vxlan_tunnel')
        assert tunnel_table == {b'src_ip': b'34.53.1.0'}
        logging.info("Tunnel table is %s", tunnel_table)

    def test_delete_config_tunnel_decap_tunnel_type(self, setup_restapi_client):
        _, _, configdb, restapi_client = setup_restapi_client
        restapi_client.post_generic_vxlan_tunnel()
        r = restapi_client.delete_config_tunnel_decap_tunnel_type('vxlan')
        assert r.status_code == 204
        # The delete is a no-op and should return 204, moreover the tunnel should not be deleted 
        tunnel_table = configdb.hgetall(VXLAN_TUNNEL_TB + '|default_vxlan_tunnel')
        assert tunnel_table == {b'src_ip': b'34.53.1.0'}


    # Encap
    def test_post_encap(self, setup_restapi_client):
        _, _, configdb, restapi_client = setup_restapi_client
        r = restapi_client.post_config_tunnel_encap_vxlan_vnid(101, None)
        assert r.status_code == 204
        keys = configdb.keys()
        assert keys == []

    def test_get_encap(self, setup_restapi_client):
        _, _, _, restapi_client = setup_restapi_client
        r = restapi_client.get_config_tunnel_encap_vxlan_vnid(101)
        assert r.status_code == 204

    def test_delete_encap(self, setup_restapi_client):
        _, _, _, restapi_client = setup_restapi_client
        r = restapi_client.delete_config_tunnel_encap_vxlan_vnid(101)
        assert r.status_code == 204


    # Vrouter
    def test_post_vrouter(self, setup_restapi_client):
        _, _, configdb, restapi_client = setup_restapi_client
        restapi_client.post_generic_vxlan_tunnel()
        r = restapi_client.post_config_vrouter_vrf_id("vnet-guid-1", {
            'vnid': 1001
        })
        assert r.status_code == 204

        vrouter_table = configdb.hgetall(VNET_TB + '|' + VNET_NAME_PREF + '1')
        assert vrouter_table == {
							b'vxlan_tunnel': b'default_vxlan_tunnel',
							b'vni': b'1001',
							b'guid': b'vnet-guid-1'
							}

    def test_post_vrouter_duplicate(self, setup_restapi_client):
        _, _, configdb, restapi_client = setup_restapi_client
        restapi_client.post_generic_vxlan_tunnel()
        r = restapi_client.post_config_vrouter_vrf_id("vnet-guid-1", {
            'vnid': 1001
        })
        assert r.status_code == 204

        vrouter_table = configdb.hgetall(VNET_TB + '|' + VNET_NAME_PREF + '1')
        assert vrouter_table == {
							b'vxlan_tunnel': b'default_vxlan_tunnel',
							b'vni': b'1001',
							b'guid': b'vnet-guid-1'
							}

        r = restapi_client.post_config_vrouter_vrf_id("vnet-guid-2", {
            'vnid': 1001
        })
        assert r.status_code == 409
        assert r.json()['error']['message'] == "Object already exists: vni=1001 vnet_name=vnet-guid-1"

    def test_post_vrouter_default(self, setup_restapi_client):
        _, _, configdb, restapi_client = setup_restapi_client
        restapi_client.post_generic_vxlan_tunnel()
        r = restapi_client.post_config_vrouter_vrf_id("Vnet-default", {
            'vnid': 2001
        })
        assert r.status_code == 204

        vrouter_table = configdb.hgetall(VNET_TB + '|' + VNET_NAME_PREF + '1')
        assert vrouter_table == {
							b'vxlan_tunnel': b'default_vxlan_tunnel',
							b'vni': b'2001',
							b'guid': b'Vnet-default',
                            b'scope': b'default'
							}

    def test_get_vrouter(self, setup_restapi_client):
        _, _, _, restapi_client = setup_restapi_client
        restapi_client.post_generic_vrouter_and_deps()
        self.check_vrouter_exists(restapi_client, "vnet-guid-1",1001)

    def test_duplicate_vni(self, setup_restapi_client):
        _, _, _, restapi_client = setup_restapi_client
        restapi_client.post_generic_vrouter_and_deps_duplicate()
        self.check_vrouter_exists(restapi_client, "vnet-guid-1",1001)

    def test_delete_vrouter(self, setup_restapi_client):
        _, _, configdb, restapi_client = setup_restapi_client
        restapi_client.post_generic_vrouter_and_deps()
        r = restapi_client.delete_config_vrouter_vrf_id("vnet-guid-1")
        assert r.status_code == 204
        vrouter_table = configdb.hgetall(VNET_TB + '|' + VNET_NAME_PREF + '1')
        assert vrouter_table == {}

    def test_guid_persistence(self, setup_restapi_client):
        _, _, _, restapi_client = setup_restapi_client
        restapi_client.post_generic_vrouter_and_deps()
        r = restapi_client.post_config_vrouter_vrf_id("vnet-guid-2", { 'vnid': 1002 })
        assert r.status_code == 204
        r = restapi_client.post_config_vrouter_vrf_id("vnet-guid-3", { 'vnid': 1003 })
        assert r.status_code == 204

        restapi_client.post_config_restart_in_mem_db()
        
        self.check_vrouter_exists(restapi_client, "vnet-guid-1",1001)
        self.check_vrouter_exists(restapi_client, "vnet-guid-2",1002)
        self.check_vrouter_exists(restapi_client, "vnet-guid-3",1003)

    def test_vnet_name_mapping_logic(self, setup_restapi_client):
        _, _, configdb, restapi_client = setup_restapi_client
        restapi_client.post_generic_vxlan_tunnel()
        for i in range (1,4):
             r = restapi_client.post_config_vrouter_vrf_id("vnet-guid-"+str(i), {'vnid': 1000+i})
             assert r.status_code == 204
             self.check_vrouter_exists(restapi_client, "vnet-guid-"+str(i), 1000+i)
             vrouter_table = configdb.hgetall(VNET_TB + '|' + VNET_NAME_PREF +str(i))
             assert vrouter_table == {
                     b'vxlan_tunnel': b'default_vxlan_tunnel',
                     b'vni': b'100'+str(i),
                     b'guid': b'vnet-guid-'+str(i)
                     }

        for i in range (1,4):
             r = restapi_client.delete_config_vrouter_vrf_id("vnet-guid-"+str(i))
             assert r.status_code == 204
             r = restapi_client.post_config_vrouter_vrf_id("vnet-guid-"+str(i+3), {'vnid': 1003+i})
             assert r.status_code == 204
             self.check_vrouter_exists(restapi_client, "vnet-guid-"+str(i+3), 1003+i)
             vrouter_table = configdb.hgetall(VNET_TB + '|' + VNET_NAME_PREF +str(i))
             assert vrouter_table == {
                     b'vxlan_tunnel': b'default_vxlan_tunnel',
                     b'vni': b'100'+str(i+3),
                     b'guid': b'vnet-guid-'+str(i+3)
                     }

             r = restapi_client.post_config_vrouter_vrf_id("vnet-guid-"+str(i+6), {'vnid': 1006+i})
             assert r.status_code == 204
             self.check_vrouter_exists(restapi_client, "vnet-guid-"+str(i+6), 1006+i)
             vrouter_table = configdb.hgetall(VNET_TB + '|' + VNET_NAME_PREF +str(i+3))
             assert vrouter_table == {
                     b'vxlan_tunnel': b'default_vxlan_tunnel',
                     b'vni': b'100'+str(i+6),
                     b'guid': b'vnet-guid-'+str(i+6)
                     }
             

    # Vlan
    def test_vlan_wo_ippref_vnetid_all_verbs(self, setup_restapi_client):
        _, _, configdb, restapi_client = setup_restapi_client
        # post
        r = restapi_client.post_config_vlan(2, {})
        assert r.status_code == 204
        
        # get
        r = restapi_client.get_config_vlan(2)
        assert r.status_code == 200
        j = json.loads(r.text)
        assert j == {
            'vlan_id': 2,
            'attr': {}
        }
        vlan_table = configdb.hgetall(VLAN_TB + '|' + VLAN_NAME_PREF + '2')
        assert vlan_table == {'host_ifname': 'MonVlan2', b'vlanid': b'2'}
        vlan_intf_table = configdb.hgetall(VLAN_INTF_TB + '|' + VLAN_NAME_PREF + '2')
        assert vlan_intf_table == {}

        # delete
        r = restapi_client.delete_config_vlan(2)
        assert r.status_code == 204
        vlan_table = configdb.hgetall(VLAN_TB + '|' + VLAN_NAME_PREF + '2')
        assert vlan_table == {}

    def test_vlan_with_vnetid_all_verbs(self, setup_restapi_client):
        _, _, configdb, restapi_client = setup_restapi_client
        # post
        restapi_client.post_generic_vrouter_and_deps()
        r = restapi_client.post_config_vlan(2, {'vnet_id' : 'vnet-guid-1'})
        assert r.status_code == 204
        
        # get
        r = restapi_client.get_config_vlan(2)
        assert r.status_code == 200
        j = json.loads(r.text)
        assert j == {
            'vlan_id': 2,
            'attr': {'vnet_id':'vnet-guid-1'}
        }
        vlan_table = configdb.hgetall(VLAN_TB + '|' + VLAN_NAME_PREF + '2')
        assert vlan_table == {'host_ifname': 'MonVlan2', b'vlanid': b'2'}
        vlan_intf_table = configdb.hgetall(VLAN_INTF_TB + '|' + VLAN_NAME_PREF + '2')
        assert vlan_intf_table == {b'proxy_arp': b'enabled', b'vnet_name': VNET_NAME_PREF + '1'}

        # delete
        r = restapi_client.delete_config_vlan(2)
        assert r.status_code == 204
        vlan_table = configdb.hgetall(VLAN_TB + '|' + VLAN_NAME_PREF + '2')
        assert vlan_table == {}
        vlan_intf_table = configdb.hgetall(VLAN_INTF_TB + '|' + VLAN_NAME_PREF + '2')
        assert vlan_intf_table == {}

    def test_vlan_with_ippref_all_verbs(self, setup_restapi_client):
        _, _, configdb, restapi_client = setup_restapi_client
        # post
        restapi_client.post_generic_vrouter_and_deps()
        r = restapi_client.post_config_vlan(2, {'ip_prefix':'10.0.1.1/24'})
        assert r.status_code == 204

        # get
        r = restapi_client.get_config_vlan(2)
        assert r.status_code == 200
        j = json.loads(r.text)
        assert j == {
            'vlan_id': 2,
            'attr': {'ip_prefix':'10.0.1.1/24'}
        }
        vlan_table = configdb.hgetall(VLAN_TB + '|' + VLAN_NAME_PREF + '2')
        assert vlan_table == {'host_ifname': 'MonVlan2', b'vlanid': b'2'}
        vlan_intf_table = configdb.hgetall(VLAN_INTF_TB + '|' + VLAN_NAME_PREF + '2|10.0.1.1/24')
        assert vlan_intf_table == {b'':b''}
        vlan_intf_table = configdb.hgetall(VLAN_INTF_TB + '|' + VLAN_NAME_PREF + '2')
        assert vlan_intf_table == {}

        # delete
        r = restapi_client.delete_config_vlan(2)
        assert r.status_code == 204
        vlan_table = configdb.hgetall(VLAN_TB + '|' + VLAN_NAME_PREF + '2')
        assert vlan_table == {}
        vlan_intf_table = configdb.hgetall(VLAN_INTF_TB + '|' + VLAN_NAME_PREF + '2|10.0.1.1/24')
        assert vlan_intf_table == {}

    def test_vlan_all_args_all_verbs(self, setup_restapi_client):
        _, _, configdb, restapi_client = setup_restapi_client
        # post
        restapi_client.post_generic_vrouter_and_deps()
        r = restapi_client.post_config_vlan(2, {'vnet_id' : 'vnet-guid-1', 'ip_prefix':'10.0.1.1/24'})
        assert r.status_code == 204

        # get
        r = restapi_client.get_config_vlan(2)
        assert r.status_code == 200
        j = json.loads(r.text)
        assert j == {
            'vlan_id': 2,
            'attr': {'vnet_id' : 'vnet-guid-1', 'ip_prefix':'10.0.1.1/24'}
        }
        vlan_table = configdb.hgetall(VLAN_TB + '|' + VLAN_NAME_PREF + '2')
        assert vlan_table == {'host_ifname': 'MonVlan2', b'vlanid': b'2'}
        vlan_intf_table = configdb.hgetall(VLAN_INTF_TB + '|' + VLAN_NAME_PREF + '2|10.0.1.1/24')
        assert vlan_intf_table == {b'':b''}
        vlan_intf_table = configdb.hgetall(VLAN_INTF_TB + '|' + VLAN_NAME_PREF + '2')
        assert vlan_intf_table == {b'proxy_arp': b'enabled', b'vnet_name': VNET_NAME_PREF+'1'}

        # delete
        r = restapi_client.delete_config_vlan(2)
        assert r.status_code == 204
        vlan_table = configdb.hgetall(VLAN_TB + '|' + VLAN_NAME_PREF + '2')
        assert vlan_table == {}
        vlan_intf_table = configdb.hgetall(VLAN_INTF_TB + '|' + VLAN_NAME_PREF + '2')
        assert vlan_intf_table == {}
        vlan_intf_table = configdb.hgetall(VLAN_INTF_TB + '|' + VLAN_NAME_PREF + '2|10.0.1.1/24')
        assert vlan_intf_table == {}

    def test_get_vlans_per_vnetid_1digitvlans(self, setup_restapi_client):
        _, _, _, restapi_client = setup_restapi_client
        # create vxlan tunnel
        restapi_client.post_config_tunnel_decap_tunnel_type('vxlan', {
        'ip_addr': '6.6.6.6'
        })
        # create vnet_id/vrf
        restapi_client.post_config_vrouter_vrf_id('vnet-guid-1', {'vnid': 1001})
        restapi_client.post_config_vrouter_vrf_id('vnet-guid-2', {'vnid': 2001})
        #create vlan interfaces
        restapi_client.post_config_vlan(3, {'vnet_id' : 'vnet-guid-1', 'ip_prefix':'10.0.4.1/24'})
        restapi_client.post_config_vlan(4, {'vnet_id' : 'vnet-guid-1', 'ip_prefix':'10.0.3.1/24'})

        restapi_client.post_config_vlan(5, {'vnet_id' : 'vnet-guid-2', 'ip_prefix':'10.2.4.1/24'})
        restapi_client.post_config_vlan(6, {'vnet_id' : 'vnet-guid-2', 'ip_prefix':'10.2.3.1/24'})
        # get vlans for vnet-guid-1
        r_vnet1 = restapi_client.get_config_interface_vlans('vnet-guid-1')
        r_vnet2 = restapi_client.get_config_interface_vlans('vnet-guid-2')
        j_vnet1 = json.loads(r_vnet1.text)
        j_vnet2 = json.loads(r_vnet2.text)
        k_vnet1 = {"vnet_id":"vnet-guid-1","attr":[{"vlan_id":3,"ip_prefix":"10.0.4.1/24"},{"vlan_id":4,"ip_prefix":"10.0.3.1/24"}]}
        k_vnet2 = {"vnet_id":"vnet-guid-2","attr":[{"vlan_id":5,"ip_prefix":"10.2.4.1/24"},{"vlan_id":6,"ip_prefix":"10.2.3.1/24"}]}
        for key,value in j_vnet1.iteritems():
            if type(value)!=list:
                #print("not type list",value)
                assert k_vnet1[key] == j_vnet1[key]
            else:
                #print("is type list",value)
                assert sorted(value) == sorted(k_vnet1.values()[0])
        for key,value in j_vnet2.iteritems():
            if type(value)!=list:
                #print("not type list",value)
                assert k_vnet2[key] == j_vnet2[key]
            else:
                #print("is type list",value)
                assert sorted(value) == sorted(k_vnet2.values()[0])

    def test_get_vlans_per_vnetid_4digitvlans(self, setup_restapi_client):
        _, _, _, restapi_client = setup_restapi_client
        # create vxlan tunnel
        restapi_client.post_config_tunnel_decap_tunnel_type('vxlan', {
        'ip_addr': '6.6.6.6'
        })
        # create vnet_id/vrf
        restapi_client.post_config_vrouter_vrf_id('vnet-guid-1', {'vnid': 1001})
        restapi_client.post_config_vrouter_vrf_id('vnet-guid-2', {'vnid': 2002})
        #create vlan interfaces
        restapi_client.post_config_vlan(1111, {'vnet_id' : 'vnet-guid-1', 'ip_prefix':'10.0.1.1/24'})
        restapi_client.post_config_vlan(2222, {'vnet_id' : 'vnet-guid-1', 'ip_prefix':'10.0.2.1/24'})
        restapi_client.post_config_vlan(3000, {'vnet_id' : 'vnet-guid-2'})
        restapi_client.post_config_vlan(4000, {'vnet_id' : 'vnet-guid-2', 'ip_prefix':'10.2.2.1/24'})

        # get vlans for vnet-guid-1
        r = restapi_client.get_config_interface_vlans('vnet-guid-1')
        j = json.loads(r.text)
        r2 = restapi_client.get_config_interface_vlans('vnet-guid-2')
        j2 = json.loads(r2.text)
        k = {"vnet_id":"vnet-guid-1","attr":[{"vlan_id":1111,"ip_prefix":"10.0.1.1/24"},{"vlan_id":2222,"ip_prefix":"10.0.2.1/24"}]}
        k2 = {"vnet_id":"vnet-guid-2","attr":[{"vlan_id":3000},{"vlan_id":4000,"ip_prefix":"10.2.2.1/24"}]}
        for key,value in j.iteritems():
            if type(value)!=list:
                #print("not type list",value)
                assert k[key] == j[key]
            else:
                #print("is type list",value)
                assert sorted(value) == sorted(k.values()[0]) 
        for key,value in j2.iteritems():
            if type(value)!=list:
                #print("not type list",value)
                assert k2[key] == j2[key]
            else:
                #print("is type list",value)
                assert sorted(value) == sorted(k2.values()[0])

    # Vlan Get
    def test_get_all_vlans(self, setup_restapi_client):
        _, _, _, restapi_client = setup_restapi_client
        # create vxlan tunnel
        restapi_client.post_config_tunnel_decap_tunnel_type('vxlan', {
        'ip_addr': '6.6.6.6'
        })
        # create vnet_id/vrf
        restapi_client.post_config_vrouter_vrf_id('vnet-guid-1', {'vnid': 1001})
        #create vlan interfaces
        restapi_client.post_config_vlan(3000, {'vnet_id' : 'vnet-guid-1', 'ip_prefix':'10.0.1.1/24'})
        restapi_client.post_config_vlan(3001, {'vnet_id' : 'vnet-guid-1'})

        # get all vlans
        r = restapi_client.get_config_vlans_all()
        j = json.loads(r.text)
        k = {"attr":[{"vlan_id":3000,"ip_prefix":"10.0.1.1/24","vnet_id":"vnet-guid-1"},{"vlan_id":3001,"vnet_id":"vnet-guid-1"}]}
        for key,value in j.iteritems():
            if type(value)!=list:
                assert k[key] == j[key]
                return
            for item in k[key]:
                if item not in value:
                    assert False

    # Vlan Member
    def test_vlan_member_tagged_untagged_interop(self, setup_restapi_client):
        _, _, _, restapi_client = setup_restapi_client
        vlan0 = 2
        vlans = [3,4]
        members = ["Ethernet2", "Ethernet3", "Ethernet4"]
        restapi_client.post_generic_vrouter_and_deps()
        r = restapi_client.post_config_vlan(vlan0, {'vnet_id' : 'vnet-guid-1', 'ip_prefix':'10.0.1.1/24'})
        assert r.status_code == 204
        for member in members:
           r = restapi_client.post_config_vlan_member(vlan0, member, {'tagging_mode' : 'untagged'})
           assert r.status_code == 204
           r = restapi_client.get_config_vlan_member(vlan0,  member)
           assert r.status_code == 200
           j = json.loads(r.text)
           assert j == {
                      'vlan_id': vlan0,
                      'if_name': member,
                      'attr': {'tagging_mode' : 'untagged'}
           }
        for vlan in vlans:
             r = restapi_client.post_config_vlan(vlan, {'vnet_id' : 'vnet-guid-1', 'ip_prefix':'10.0.1.1/24'})
             assert r.status_code == 204
             # post
             for member in members:
                 r = restapi_client.post_config_vlan_member(vlan, member, {'tagging_mode' : 'tagged'})
                 assert r.status_code == 204

                 # get
                 r = restapi_client.get_config_vlan_member(vlan,  member)
                 assert r.status_code == 200
                 j = json.loads(r.text)
                 assert j == {
                      'vlan_id': vlan,
                      'if_name': member,
                      'attr': {'tagging_mode' : 'tagged'}
                 }

    def test_vlan_member_tagging_all_verbs(self, setup_restapi_client):
        _, _, configdb, restapi_client = setup_restapi_client
        vlans = [2,3]
        members = ['Ethernet2', 'Ethernet3']
        restapi_client.post_generic_vrouter_and_deps()
        for vlan in vlans:
             r = restapi_client.post_config_vlan(vlan, {'vnet_id' : 'vnet-guid-1', 'ip_prefix':'10.0.1.1/24'})
             assert r.status_code == 204
             # post
             for member in members:
                 r = restapi_client.post_config_vlan_member(vlan, member, {'tagging_mode' : 'tagged'})
                 assert r.status_code == 204

                 # get
                 r = restapi_client.get_config_vlan_member(vlan,  member)
                 assert r.status_code == 200
                 j = json.loads(r.text)
                 assert j == {
                      'vlan_id': vlan,
                      'if_name': member,
                      'attr': {'tagging_mode' : 'tagged'}
                 }
                 vlan_mem_table = configdb.hgetall(VLAN_MEMB_TB + '|' + VLAN_NAME_PREF +str(vlan)+'|'+member)
                 assert vlan_mem_table == {b'tagging_mode':b'tagged'}

                 # delete
                 r = restapi_client.delete_config_vlan_member(vlan, member)
                 assert r.status_code == 204
                 vlan_mem_table = configdb.hgetall(VLAN_MEMB_TB + '|' + VLAN_NAME_PREF + str(vlan) + "|" +  member)
                 assert vlan_mem_table == {}

    def test_vlan_member_notagging_all_verbs(self, setup_restapi_client):
        _, _, configdb, restapi_client = setup_restapi_client
        # post
        restapi_client.post_generic_vlan_and_deps()
        r = restapi_client.post_config_vlan_member(2, "Ethernet2", {})
        assert r.status_code == 204

        # get
        r = restapi_client.get_config_vlan_member(2, "Ethernet2")
        assert r.status_code == 200
        j = json.loads(r.text)
        assert j == {
            'vlan_id': 2,
            'if_name': 'Ethernet2',
            'attr': {'tagging_mode' : 'untagged'}
        }
        vlan_mem_table = configdb.hgetall(VLAN_MEMB_TB + '|' + VLAN_NAME_PREF + '2|Ethernet2')
        assert vlan_mem_table == {b'tagging_mode':b'untagged'}

    def test_get_members_per_vlan(self, setup_restapi_client):
        _, _, _, restapi_client = setup_restapi_client
        # create vxlan tunnel
        restapi_client.post_config_tunnel_decap_tunnel_type('vxlan', {
        'ip_addr': '6.6.6.6'
        })
        # create vnet_id/vrf
        restapi_client.post_config_vrouter_vrf_id('vnet-guid-1', {'vnid': 1001})
        #create vlan interface 2
        restapi_client.post_config_vlan(2, {'vnet_id' : 'vnet-guid-1', 'ip_prefix':'10.0.1.1/24'})
        members = ["Ethernet2", "Ethernet3", "Ethernet4"]
        for member in members:
            restapi_client.post_config_vlan_member(2, member, {'tagging_mode' : 'untagged'})
        r = restapi_client.get_config_interface_vlan_members(2)
        j = json.loads(r.text)
        assert sorted(j) == sorted(
            {"vlan_id":2,"attr":[{"if_name":"Ethernet2","tagging_mode":"untagged"},{"if_name":"Ethernet3","tagging_mode":"untagged"},{"if_name":"Ethernet4","tagging_mode":"untagged"}]}
            )


    # Vlan Neighbor
    def test_vlan_neighbor_all_verbs(self, setup_restapi_client):
        _, _, configdb, restapi_client = setup_restapi_client
        # post
        restapi_client.post_generic_vlan_and_deps()
        r = restapi_client.post_config_vlan_neighbor(2, "10.10.10.10")
        assert r.status_code == 204

        # get
        r = restapi_client.get_config_vlan_neighbor(2, "10.10.10.10")
        assert r.status_code == 200
        j = json.loads(r.text)
        assert j == {
            'vlan_id': 2,
            'ip_addr': '10.10.10.10'
        }
        vlan_neigh_table = configdb.hgetall(VLAN_NEIGH_TB + '|' + VLAN_NAME_PREF + '2|10.10.10.10')
        assert vlan_neigh_table == {b'family':b'IPv4'}

        # delete
        r = restapi_client.delete_config_vlan_neighbor(2, "10.10.10.10")
        assert r.status_code == 204
        vlan_neigh_table = configdb.hgetall(VLAN_NEIGH_TB + '|' + VLAN_NAME_PREF + '2|10.10.10.10')
        assert vlan_neigh_table == {}

    def test_get_neighbors_per_vlan(self, setup_restapi_client):
        _, _, _, restapi_client = setup_restapi_client
        # create vxlan tunnel
        restapi_client.post_config_tunnel_decap_tunnel_type('vxlan', {
        'ip_addr': '6.6.6.6'
        })
        # create vnet_id/vrf
        restapi_client.post_config_vrouter_vrf_id('vnet-guid-1', {'vnid': 1001})
        #create vlan interface 2
        restapi_client.post_config_vlan(3, {'vnet_id' : 'vnet-guid-1', 'ip_prefix':'10.0.2.1/24'})
        restapi_client.post_config_vlan_neighbor(3, "10.10.20.10")
        restapi_client.post_config_vlan_neighbor(3, "10.10.30.10")

        # get vlans for vnet-guid-1
        r = restapi_client.get_config_interface_vlan_neighbors(3)
        j = json.loads(r.text)
        assert sorted(j) == sorted(
            {"vlan_id":3,"attr":[{"ip_addr":"10.10.20.10"},{"ip_addr":"10.10.30.10"}]}
            )

    # Routes
    def test_patch_update_routes_with_optional_args(self, setup_restapi_client):
        db, _, _, restapi_client = setup_restapi_client
        restapi_client.post_generic_vlan_and_deps()
        # No optional args
        route = {
                    'cmd':'add',
                    'ip_prefix':'10.2.1.0/24',
                    'nexthop':'192.168.2.1'
                }
        r = restapi_client.patch_config_vrouter_vrf_id_routes("vnet-guid-1", [route])
        assert r.status_code == 204
        route_table = db.hgetall(ROUTE_TUN_TB + ':' + VNET_NAME_PREF +str(1)+':'+route['ip_prefix'])
        assert route_table == {b'endpoint' : route['nexthop']}
        del route['cmd']
        routes = list()
        routes.append(route)
        routes.append({'nexthop': '', 'ip_prefix': '10.1.1.0/24', 'ifname': 'Vlan2'})
        r = restapi_client.get_config_vrouter_vrf_id_routes("vnet-guid-1")
        assert r.status_code == 200
        j = json.loads(r.text)
        assert sorted(j) == sorted(routes)

        # Vnid Optional arg
        route['vnid'] = 5000
        route['cmd'] = 'add'
        r = restapi_client.patch_config_vrouter_vrf_id_routes("vnet-guid-1", [route])
        assert r.status_code == 204
        route_table = db.hgetall(ROUTE_TUN_TB + ':' + VNET_NAME_PREF +str(1)+':'+route['ip_prefix'])
        assert route_table == {b'endpoint' : route['nexthop'],
                                       b'vni' : str(route['vnid'])
                                      }
        del route['cmd']
        routes = list()
        routes.append(route)
        routes.append({'nexthop': '', 'ip_prefix': '10.1.1.0/24', 'ifname': 'Vlan2'})
        r = restapi_client.get_config_vrouter_vrf_id_routes("vnet-guid-1")
        assert r.status_code == 200
        j = json.loads(r.text)
        assert sorted(j) == sorted(routes)       

        # Mac address Optional arg
        del route['vnid']
        route['mac_address'] = '00:08:aa:bb:cd:ef'
        route['cmd'] = 'add'
        r = restapi_client.patch_config_vrouter_vrf_id_routes("vnet-guid-1", [route])
        assert r.status_code == 204
        route_table = db.hgetall(ROUTE_TUN_TB + ':' + VNET_NAME_PREF +str(1)+':'+route['ip_prefix'])
        assert route_table == {b'endpoint' : route['nexthop'],
                                       b'mac_address' : route['mac_address']
                                      }
        del route['cmd']
        routes = list()
        routes.append(route)
        routes.append({'nexthop': '', 'ip_prefix': '10.1.1.0/24', 'ifname': 'Vlan2'})
        r = restapi_client.get_config_vrouter_vrf_id_routes("vnet-guid-1")
        assert r.status_code == 200
        j = json.loads(r.text)
        assert sorted(j) == sorted(routes)

        # Weight optional arg
        route['vnid'] = 5000
        route['nexthop'] = '100.3.152.32,200.3.152.32'
        route['weight'] = '20,10'
        route['cmd'] = 'add'
        r = restapi_client.patch_config_vrouter_vrf_id_routes("vnet-guid-1", [route])
        assert r.status_code == 204
        route_table = db.hgetall(ROUTE_TUN_TB + ':' + VNET_NAME_PREF +str(1)+':'+route['ip_prefix'])
        assert route_table == {b'endpoint' : route['nexthop'],
                                       b'vni': str(route['vnid']), 
                                       b'mac_address' : route['mac_address'],
                                       b'weight': route['weight']
                                      }
        del route['cmd']
        routes = list()
        routes.append(route)
        routes.append({'nexthop': '', 'ip_prefix': '10.1.1.0/24', 'ifname': 'Vlan2'})
        r = restapi_client.get_config_vrouter_vrf_id_routes("vnet-guid-1")
        assert r.status_code == 200
        j = json.loads(r.text)
        assert sorted(j) == sorted(routes)
        del route['weight']

        # Profile optional arg
        route['vnid'] = 5000
        route['nexthop'] = '100.3.152.32,200.3.152.32'
        route['profile'] = 'profile1'
        route['cmd'] = 'add'
        r = restapi_client.patch_config_vrouter_vrf_id_routes("vnet-guid-1", [route])
        assert r.status_code == 204
        route_table = db.hgetall(ROUTE_TUN_TB + ':' + VNET_NAME_PREF +str(1)+':'+route['ip_prefix'])
        assert route_table == {b'endpoint' : route['nexthop'],
                                       b'vni': str(route['vnid']), 
                                       b'mac_address' : route['mac_address'],
                                       b'profile': route['profile']
                                      }
        del route['cmd']
        routes = list()
        routes.append(route)
        routes.append({'nexthop': '', 'ip_prefix': '10.1.1.0/24', 'ifname': 'Vlan2'})
        r = restapi_client.get_config_vrouter_vrf_id_routes("vnet-guid-1")
        assert r.status_code == 200
        j = json.loads(r.text)
        assert sorted(j) == sorted(routes)     


    def test_patch_routes_drop_bm_routes_tunnel(self, setup_restapi_client):
        _, _, _, restapi_client = setup_restapi_client
        restapi_client.post_generic_vrouter_and_deps()
        rv = restapi_client.post_config_vlan(2, {'vnet_id' : 'vnet-guid-1', 'ip_prefix' : '10.1.1.0/24'})
        assert rv.status_code == 204
        r = restapi_client.post_config_vlan(3, {'vnet_id' : 'vnet-guid-1', 'ip_prefix':'10.1.5.0/24'})
        assert r.status_code == 204
        routes = []
        for i in range (1,7):
            routes.append({'cmd':'add',
                            'ip_prefix':'10.5.'+str(i)+'.0/'+str(24),
                            'nexthop':'34.53.'+str(i)+'.0',
                            'vnid': 1 + i%5,
                            'mac_address':'00:08:aa:bb:cd:'+hex(15+i)[2:]})
            routes.append({'cmd':'add',
                            'ip_prefix':'10.5.'+str(i)+'.0/'+str(30),
                            'nexthop':'34.53.'+str(i)+'.0',
                            'vnid': 1 + i%5,
                            'mac_address':'00:08:aa:bb:cd:'+hex(15+i)[2:]})
            routes.append({'cmd':'add',
                            'ip_prefix':'10.5.'+str(i)+'.1/'+str(32),
                            'nexthop':'34.53.'+str(i)+'.0',
                            'vnid': 1 + i%5,
                            'mac_address':'00:08:aa:bb:cd:'+hex(15+i)[2:]})

        r = restapi_client.patch_config_vrouter_vrf_id_routes("vnet-guid-1", routes)
        assert r.status_code == 204

        routes_bm = []
        routes_not_bm = []
        for route in routes:
             del route['cmd']
             if route['nexthop'] == '34.53.1.0':
                  routes_bm.append(route)
             else:
                  routes_not_bm.append(route)
        restapi_client.check_routes_exist_in_tun_tb(1, routes_not_bm) 
        restapi_client.check_routes_dont_exist_in_tun_tb(1, routes_bm) 

        routes_not_bm.append({'nexthop': '', 'ip_prefix': '10.1.1.0/24', 'ifname': 'Vlan2'})
        routes_not_bm.append({'nexthop': '', 'ip_prefix': '10.1.5.0/24', 'ifname': 'Vlan3'})
        r = restapi_client.get_config_vrouter_vrf_id_routes("vnet-guid-1")
        assert r.status_code == 200
        j = json.loads(r.text)
        assert sorted(j) == sorted(routes_not_bm)

        for route in routes_bm:
             route['cmd'] = 'delete'
        r = restapi_client.patch_config_vrouter_vrf_id_routes("vnet-guid-1", routes_bm)
        assert r.status_code == 204

    def test_patch_routes_drop_bm_routes_local(self, setup_restapi_client):
        _, _, _, restapi_client = setup_restapi_client
        restapi_client.post_generic_vrouter_and_deps()
        rv = restapi_client.post_config_vlan(2, {'vnet_id':'vnet-guid-1', 'ip_prefix':'10.1.1.0/24'})
        assert rv.status_code == 204
        r = restapi_client.post_config_vlan(3, {'vnet_id':'vnet-guid-1', 'ip_prefix':'10.1.5.0/24'})
        assert r.status_code == 204
        routes = []
        for i in range (1,7):
            routes.append({'cmd':'add',
                            'ip_prefix':'10.5.'+str(i)+'.0/'+str(24),
                            'nexthop':'34.53.'+str(i)+'.0',
                            'ifname': 'Vlan3005'})
            routes.append({'cmd':'add',
                            'ip_prefix':'10.5.'+str(i)+'.0/'+str(30),
                            'nexthop':'34.53.'+str(i)+'.0',
                            'ifname': 'Vlan3005'})
            routes.append({'cmd':'add',
                            'ip_prefix':'10.5.'+str(i)+'.1/'+str(32),
                            'nexthop':'34.53.'+str(i)+'.0',
                            'ifname': 'Vlan3005'})

        r = restapi_client.patch_config_vrouter_vrf_id_routes("vnet-guid-1", routes)
        assert r.status_code == 204

        routes_bm = []
        routes_not_bm = []
        for route in routes:
             del route['cmd']
             if route['nexthop'] == '34.53.1.0':
                  routes_bm.append(route)
             else:
                  routes_not_bm.append(route)
        restapi_client.check_routes_exist_in_loc_route_tb(1, routes_not_bm)
        restapi_client.check_routes_dont_exist_in_loc_route_tb(1, routes_bm)

        r = restapi_client.get_config_vrouter_vrf_id_routes("vnet-guid-1")
        assert r.status_code == 200
        j = json.loads(r.text)
        j.remove({'nexthop': '', 'ifname': 'Vlan3', 'ip_prefix': '10.1.5.0/24'})
        j.remove({'nexthop': '', 'ifname': 'Vlan2', 'ip_prefix': '10.1.1.0/24'})
        assert sorted(j) == sorted(routes_not_bm)
        for route in routes_bm:
             route['cmd'] = 'delete'
        r = restapi_client.patch_config_vrouter_vrf_id_routes("vnet-guid-1", routes_bm)
        assert r.status_code == 204

    def test_routes_all_verbs(self, setup_restapi_client):
        _, _, _, restapi_client = setup_restapi_client
        restapi_client.post_generic_vlan_and_deps()
        routes = []
        rv = restapi_client.post_config_vrouter_vrf_id("vnet-guid-2", {
            'vnid': 1002
        })
        assert rv.status_code == 204
        for i in range (1,100):
             routes.append({'cmd':'add', 
                            'ip_prefix':'10.2.'+str(i)+'.0/24', 
                            'nexthop':'192.168.2.'+str(i), 
                            'vnid': 1 + i%5, 
                            'mac_address':'00:08:aa:bb:cd:'+hex(15+i)[2:]})
       
        # Patch add
        r = restapi_client.patch_config_vrouter_vrf_id_routes("vnet-guid-1", routes)
        assert r.status_code == 204
        restapi_client.check_routes_exist_in_tun_tb(1, routes)
        r = restapi_client.patch_config_vrouter_vrf_id_routes("vnet-guid-2", routes)
        assert r.status_code == 204
        restapi_client.check_routes_exist_in_tun_tb(2, routes)
        # Patch delete

        # Get all
        for route in routes:
             del route['cmd']
        routes.append({'nexthop': '', 'ip_prefix': '10.1.1.0/24', 'ifname': 'Vlan2'})
        r = restapi_client.get_config_vrouter_vrf_id_routes("vnet-guid-1")
        assert r.status_code == 200
        j = json.loads(r.text)
        assert sorted(j) == sorted(routes)

        # Get filtered by vnid
        routes_vnid = []
        routes_not_vnid = []
        route_pref = {}
        i = 0
        for route in routes:
            if i == 70:
                 route_pref = route
            if 'vnid' in route and route['vnid'] == 5:
                 routes_vnid.append(route)
            else:
                 routes_not_vnid.append(route)
            i += 1
        routes_vnid.append({'nexthop': '', 'ifname': 'Vlan2', 'ip_prefix': '10.1.1.0/24'})
        r = restapi_client.get_config_vrouter_vrf_id_routes("vnet-guid-1", vnid=5)
        assert r.status_code == 200
        j = json.loads(r.text)
        assert sorted(j) == sorted(routes_vnid)
        routes_vnid.remove({'nexthop': '', 'ifname': 'Vlan2', 'ip_prefix': '10.1.1.0/24'})

        # Get filtered by ip_prefix
        r = restapi_client.get_config_vrouter_vrf_id_routes("vnet-guid-1", ip_prefix=route_pref['ip_prefix'])
        assert r.status_code == 200
        j = json.loads(r.text)
        assert sorted(j) == sorted([route_pref])
 
        # Get filtered by both ip_prefix and vnid
        r = restapi_client.get_config_vrouter_vrf_id_routes("vnet-guid-1", vnid=2, ip_prefix=route_pref['ip_prefix'])
        assert r.status_code == 200
        j = json.loads(r.text)
        assert sorted(j) == sorted([route_pref])

        # Delete filtered by vnid
        r = restapi_client.delete_config_vrouter_vrf_id_routes("vnet-guid-1", vnid=5)
        assert r.status_code == 204
        routes_not_vnid_cleaned = routes_not_vnid
        for route in routes_not_vnid:
            if "mac_address" not in route:
                routes_not_vnid_cleaned.remove(route)
        restapi_client.check_routes_exist_in_tun_tb(1, routes_not_vnid_cleaned)
        restapi_client.check_routes_dont_exist_in_tun_tb(1, routes_vnid)
        r = restapi_client.get_config_vrouter_vrf_id_routes("vnet-guid-1", vnid=5)
        j = json.loads(r.text)
        assert j == []

        # Patch combo add and delete
        routes_cleaned = []
        for route in routes:
            if len(route["nexthop"]) > 1:
                if "vnid" in route and route['vnid'] == 5:
                    route['cmd'] = 'add'
                else:
                    route['cmd'] = 'delete'
                routes_cleaned.append(route)
        r = restapi_client.patch_config_vrouter_vrf_id_routes("vnet-guid-1", routes_cleaned)
        assert r.status_code == 204
        restapi_client.check_routes_exist_in_tun_tb(1, routes_vnid)
        restapi_client.check_routes_dont_exist_in_tun_tb(1, routes_not_vnid)
        for route in routes_cleaned:
             del route['cmd']
        r = restapi_client.get_config_vrouter_vrf_id_routes("vnet-guid-1")
        assert r.status_code == 200
        j = json.loads(r.text)
        assert sorted(j) == sorted(routes_vnid)
        r = restapi_client.get_config_vrouter_vrf_id_routes("vnet-guid-1", vnid=4)
        assert r.status_code == 200
        j = json.loads(r.text)
        assert j == []

        # Delete all routes
        r = restapi_client.delete_config_vrouter_vrf_id_routes("vnet-guid-1")
        assert r.status_code == 204
        restapi_client.check_routes_dont_exist_in_tun_tb(1, routes)
        r = restapi_client.get_config_vrouter_vrf_id_routes("vnet-guid-1")
        j = json.loads(r.text)
        assert j == []

        # Test that routes in other Vnet are untouched
        restapi_client.check_routes_exist_in_tun_tb(2, routes_cleaned)
        r = restapi_client.get_config_vrouter_vrf_id_routes("vnet-guid-2")
        assert r.status_code == 200
        j = json.loads(r.text)
        assert sorted(j) == sorted(routes_cleaned)
        
    def test_vrf_routes_all_verbs(self, setup_restapi_client):
        _, _, _, restapi_client = setup_restapi_client
        routes = []
        routes.append({'cmd':'add',
                            'ip_prefix':'20.1.2.0/24',
                            'nexthop':'192.168.2.200'})

        routes.append({'cmd':'add',
                            'ip_prefix':'30.1.2.0/24',
                            'nexthop':'192.168.2.200,192.168.2.201'})

        routes.append({'cmd':'add',
                            'ip_prefix':'40.1.2.0/24',
                            'nexthop':'192.168.2.200,192.168.2.201,192.168.2.202',
                            'ifname':'Ethernet0,Ethernet4,Ethernet8'})

        routes.append({'cmd':'add',
                            'ip_prefix':'50.1.2.0/24',
                            'ifname':'Ethernet0,Ethernet4'})

        routes.append({'cmd':'add',
                            'ip_prefix':'60.1.2.0/24',
                            'ifname':'Ethernet8'})

        routes.append({'cmd':'add',
                            'ip_prefix':'70.1.2.0/24',
                            'nexthop':'192.168.2.200,192.168.2.201,192.168.2.202',
                            'weight':'10,20',
                            'profile':'profile1'})


        # Patch add
        r = restapi_client.patch_config_vrf_vrf_id_routes("default", routes)
        assert r.status_code == 204

        for route in routes:
             del route['cmd']
             if 'nexthop' not in route:
                 route['nexthop'] = ''

        r = restapi_client.get_config_vrf_vrf_id_routes("default")
        assert r.status_code == 200
        j = json.loads(r.text)
        assert sorted(j) == sorted(routes)

        # Patch del
        for route in routes:
            route['cmd'] = 'delete'
            if route['nexthop'] == '':
                del route['nexthop']

        r = restapi_client.patch_config_vrf_vrf_id_routes("default", routes)
        assert r.status_code == 204

        r = restapi_client.get_config_vrf_vrf_id_routes("default")
        assert r.status_code == 200
        j = json.loads(r.text)
        routes = []
        assert sorted(j) == sorted(routes)

        # Test modify
        routes.append({'cmd':'add',
                            'ip_prefix':'40.1.2.0/24',
                            'nexthop':'192.168.2.200,192.168.2.201',
                            'ifname':'Ethernet0,Ethernet4'})

        r = restapi_client.patch_config_vrf_vrf_id_routes("default", routes)
        assert r.status_code == 204

        for route in routes:
            route['nexthop'] = '192.168.2.200,192.168.2.201,10.1.1.1'
            route['ifname'] = 'Ethernet0,Ethernet4,Vlan1000'

        r = restapi_client.patch_config_vrf_vrf_id_routes("default", routes)
        assert r.status_code == 204

        for route in routes:
            del route['cmd']

        r = restapi_client.get_config_vrf_vrf_id_routes("default")
        assert r.status_code == 200

        j = json.loads(r.text)
        assert sorted(j) == sorted(routes)

    def test_local_subnet_route_addition(self, setup_restapi_client):
        db, _, _, restapi_client = setup_restapi_client
        restapi_client.post_generic_vlan_and_deps()
        local_route_table = db.hgetall(LOCAL_ROUTE_TB + ':' + VNET_NAME_PREF +str(1)+':10.1.1.0/24')
        assert local_route_table == {b'ifname' : VLAN_NAME_PREF + '2'}
        r = restapi_client.delete_config_vlan(2)
        assert r.status_code == 204
        local_route_table = db.hgetall(LOCAL_ROUTE_TB + ':' + VNET_NAME_PREF +str(1)+':10.1.1.0/24')
        assert local_route_table == {}

    # Operations
    # PingVRF
    def test_post_ping(self, setup_restapi_client):
        _, _, _, restapi_client = setup_restapi_client
        restapi_client.post_generic_vrouter_and_deps()
        # Ping loss but response 200
        r = restapi_client.post_ping({'vnet_id' : 'vnet-guid-1', 'count' : '2', 'ip_addr' : '8.8.8.8'})
        assert r.status_code == 200
        # Ping success and response 200
        r = restapi_client.post_ping({"count" : "2", "ip_addr" : "8.8.8.8"})
        assert r.status_code == 200
        # Ping success and response 200
        r = restapi_client.post_ping({"ip_addr" : "8.8.8.8"})
        assert r.status_code == 200
        
class TestRestApiNegative():
    """Invalid input tests"""
    # Decap:
    def test_delete_config_tunnel_decap_tunnel_type_not_vxlan(self, setup_restapi_client):
        _, _, _, restapi_client = setup_restapi_client
        r = restapi_client.delete_config_tunnel_decap_tunnel_type('not_vxlan')
        assert r.status_code == 400

        j = json.loads(r.text)
        assert ['tunnel_type'] == j['error']['fields']

    def test_get_config_tunnel_decap_tunnel_not_created(self, setup_restapi_client):
        _, _, _, restapi_client = setup_restapi_client
        r = restapi_client.get_config_tunnel_decap_tunnel_type('vxlan')
        assert r.status_code == 404

        j = json.loads(r.text)
        assert ['tunnel_type'] == j['error']['fields']

    # Vrouter: 
    def test_delete_vrouter_with_dependencies(self, setup_restapi_client):
        _, _, _, restapi_client = setup_restapi_client
        # Init
        restapi_client.post_generic_vrouter_and_deps()
        r = restapi_client.post_config_vrouter_vrf_id("vnet-guid-2", { 'vnid': 1002 })
        assert r.status_code == 204
        
        # Vlan Dependency
        rv = restapi_client.post_config_vlan(2, {'vnet_id' : 'vnet-guid-1', 'ip_prefix' : '10.1.1.0/24'})
        assert rv.status_code == 204
        rv = restapi_client.post_config_vlan(3,  {'ip_prefix' : '10.1.1.0/24'})
        assert rv.status_code, 204
        r = restapi_client.delete_config_vrouter_vrf_id("vnet-guid-1")
        assert r.status_code == 409
        j = json.loads(r.text)
        assert DELETE_DEP == j['error']['sub-code']
        rv = restapi_client.delete_config_vlan(2)
        assert rv.status_code == 204
        r = restapi_client.delete_config_vrouter_vrf_id("vnet-guid-1")
        assert r.status_code == 204

        # Routes Dependency
        rv = restapi_client.post_config_vrouter_vrf_id("vnet-guid-1", {
            'vnid': 1001
        })
        assert rv.status_code == 204
        r = restapi_client.patch_config_vrouter_vrf_id_routes("vnet-guid-1", [{'cmd':'add', 'ip_prefix':'10.1.2.0/24', 'nexthop':'192.168.2.1'}])
        assert r.status_code == 204
        r = restapi_client.delete_config_vrouter_vrf_id("vnet-guid-1")
        assert r.status_code == 409
        j = json.loads(r.text)
        assert DELETE_DEP == j['error']['sub-code']
        r = restapi_client.patch_config_vrouter_vrf_id_routes("vnet-guid-1", [{'cmd':'delete', 'ip_prefix':'10.1.2.0/24', 'nexthop':'192.168.2.1'}])
        assert r.status_code == 204
        r = restapi_client.delete_config_vrouter_vrf_id("vnet-guid-1")
        assert r.status_code == 204


    def test_vrouter_not_created_all_verbs(self, setup_restapi_client):
        _, _, _, restapi_client = setup_restapi_client
        # Get
        r = restapi_client.get_config_vrouter_vrf_id("vnet-guid-1")
        assert r.status_code == 404

        # Delete
        r = restapi_client.delete_config_vrouter_vrf_id("vnet-guid-1")
        assert r.status_code == 404

        # Vrouter Routes
        r = restapi_client.patch_config_vrouter_vrf_id_routes("vnet-guid-1", [{'cmd':'add', 'ip_prefix':'10.1.2.0/24', 'nexthop':'192.168.2.1'}])
        assert r.status_code == 404
        j = json.loads(r.text)
        assert ['vnet-guid-1'] == j['error']['fields']
        r = restapi_client.get_config_vrouter_vrf_id_routes("vnet-guid-1")
        assert r.status_code == 404
        j = json.loads(r.text)
        assert ['vnet-guid-1'] == j['error']['fields']
        r = restapi_client.delete_config_vrouter_vrf_id_routes("vnet-guid-1")
        assert r.status_code == 404
        j = json.loads(r.text)
        assert ['vnet-guid-1'] == j['error']['fields']

    def test_post_vrouter_without_vtep(self, setup_restapi_client):
        _, _, _, restapi_client = setup_restapi_client
        r = restapi_client.post_config_vrouter_vrf_id("vnet-guid-1", {
            'vnid': 1001
        })
        assert r.status_code == 409

        j = json.loads(r.text)
        assert ['tunnel'] == j['error']['fields']
        assert DEP_MISSING == j['error']['sub-code']

    def test_post_vrouter_which_exists(self, setup_restapi_client):
        _, _, _, restapi_client = setup_restapi_client
        restapi_client.post_generic_vrouter_and_deps()
        r = restapi_client.post_config_vrouter_vrf_id("vnet-guid-1", {
            'vnid': 1001
        })
        assert r.status_code == 409

        j = json.loads(r.text)
        assert RESRC_EXISTS == j['error']['sub-code']

    def test_post_vrouter_malformed_arg(self, setup_restapi_client):
        _, _, _, restapi_client = setup_restapi_client
        restapi_client.post_generic_vrouter_and_deps()
        r = restapi_client.post_config_vrouter_vrf_id("vnet-guid-1", {
            'vnid': "this is malformed"
        })
        assert r.status_code == 400

        j = json.loads(r.text)
        assert ['vnid'] == j['error']['fields']

    # Vlan
    def test_post_vlan_which_exists(self, setup_restapi_client):
        _, _, _, restapi_client = setup_restapi_client
        restapi_client.post_generic_vlan_and_deps()
        r = restapi_client.post_config_vlan(2, {})
        assert r.status_code == 409

        j = json.loads(r.text)
        assert RESRC_EXISTS == j['error']['sub-code']

    def test_vlan_not_created_all_verbs(self, setup_restapi_client):
        _, _, _, restapi_client = setup_restapi_client
        # Get
        r = restapi_client.get_config_vlan(2)
        assert r.status_code == 404 
        r = restapi_client.get_config_vlan_member(2, "ethernet2")
        assert r.status_code == 404
        j = json.loads(r.text)
        assert ['vlan_id'] == j['error']['fields']
        r = restapi_client.get_config_vlan_neighbor(2, "10.10.10.10")
        assert r.status_code == 404
        j = json.loads(r.text)
        assert ['vlan_id'] == j['error']['fields']

        # Delete
        r = restapi_client.delete_config_vlan(2)
        assert r.status_code == 404
        r = restapi_client.delete_config_vlan_member(2, "ethernet2")
        assert r.status_code == 404
        j = json.loads(r.text)
        assert ['vlan_id'] == j['error']['fields']
        r = restapi_client.delete_config_vlan_neighbor(2, "10.10.10.10")
        assert r.status_code == 404
        j = json.loads(r.text)
        assert ['vlan_id'] == j['error']['fields']

        # Post
        r = restapi_client.post_config_vlan_member(2, "ethernet2", {})
        assert r.status_code == 404
        j = json.loads(r.text)
        assert ['vlan_id'] == j['error']['fields']
        r = restapi_client.post_config_vlan_neighbor(2, "10.10.10.10")
        assert r.status_code == 404
        j = json.loads(r.text)
        assert ['vlan_id'] == j['error']['fields']

    def test_vlan_out_of_range(self, setup_restapi_client):
        _, _, _, restapi_client = setup_restapi_client
        vlan_ids = [1,4095]
        member = "Ethernet1"
        ip_addr = "10.10.1.1"
        for vlan_id in vlan_ids:
                r = restapi_client.post_config_vlan(vlan_id, {})
                assert r.status_code == 400
                r = restapi_client.get_config_vlan(vlan_id)
                assert r.status_code == 400
                r = restapi_client.delete_config_vlan(vlan_id)
                assert r.status_code == 400
                
                r = restapi_client.post_config_vlan_member(vlan_id, member, {})
                assert r.status_code == 400
                r = restapi_client.get_config_vlan_member(vlan_id, member)
                assert r.status_code == 400
                r = restapi_client.delete_config_vlan_member(vlan_id, member)         
                assert r.status_code == 400

                r = restapi_client.post_config_vlan_neighbor(vlan_id, {})
                assert r.status_code == 400
                r = restapi_client.get_config_vlan_neighbor(vlan_id, ip_addr)
                assert r.status_code == 400
                r = restapi_client.delete_config_vlan_neighbor(vlan_id, ip_addr)
                assert r.status_code == 400

    def test_delete_vlan_with_dependencies(self, setup_restapi_client):
        _, _, _, restapi_client = setup_restapi_client
        # Init generic config
        restapi_client.post_generic_vlan_and_deps()
        rv = restapi_client.post_config_vlan(3, {'vnet_id' : 'vnet-guid-1', 'ip_prefix' : '10.1.1.0/24'})
        assert rv.status_code == 204
        rv = restapi_client.post_config_vlan_member(3, "Ethernet2", {})
        assert rv.status_code == 204

        # Dependency Vlan Member
        rv = restapi_client.post_config_vlan_member(2, "Ethernet1", {})
        assert rv.status_code == 204
        r = restapi_client.delete_config_vlan(2)
        assert r.status_code == 409
        j = json.loads(r.text)
        assert DELETE_DEP == j['error']['sub-code']
        rv = restapi_client.delete_config_vlan_member(2, "Ethernet1")
        assert rv.status_code == 204
        r = restapi_client.delete_config_vlan(2)
        assert r.status_code == 204

        # Dependency Vlan Neighbor
        rv = restapi_client.post_config_vlan(2, {'vnet_id' : 'vnet-guid-1', 'ip_prefix' : '10.1.1.0/24'})
        rv = restapi_client.post_config_vlan_neighbor(2, "10.10.10.10")
        assert rv.status_code == 204
        r = restapi_client.delete_config_vlan(2)
        assert r.status_code == 409
        j = json.loads(r.text)
        assert DELETE_DEP == j['error']['sub-code']
        rv = restapi_client.delete_config_vlan_neighbor(2, "10.10.10.10")
        assert rv.status_code == 204
        r = restapi_client.delete_config_vlan(2)
        assert r.status_code == 204

    def test_get_vlans_per_vnetid_invalid_vlan(self, setup_restapi_client):
        _, _, _, restapi_client = setup_restapi_client
        # create vxlan tunnel
        restapi_client.post_config_tunnel_decap_tunnel_type('vxlan', {
        'ip_addr': '6.6.6.6'
        })
        # create vnet_id/vrf
        restapi_client.post_config_vrouter_vrf_id('vnet-guid-1', {'vnid': 1001})
        #create invalid vlan interfaces
        restapi_client.post_config_vlan(5555, {'vnet_id' : 'vnet-guid-1', 'ip_prefix':'10.0.1.1/24'})
        restapi_client.post_config_vlan(4096, {'vnet_id' : 'vnet-guid-1', 'ip_prefix':'10.0.2.1/24'})

        # get vlans for vnet-guid-1
        r = restapi_client.get_config_interface_vlans('vnet-guid-1')
        j = json.loads(r.text)
        assert j == {u'attr': None, u'vnet_id': u'vnet-guid-1'}

    def test_get_vlans_per_vnetid_invalid_vnet(self, setup_restapi_client):
        _, _, _, restapi_client = setup_restapi_client
        # create vxlan tunnel
        restapi_client.post_config_tunnel_decap_tunnel_type('vxlan', {
        'ip_addr': '6.6.6.6'
        })
        # create vnet_id/vrf
        restapi_client.post_config_vrouter_vrf_id('vnet-guid-1', {'vnid': 1001})
        #create vlan interfaces
        restapi_client.post_config_vlan(555, {'vnet_id' : 'vnet-guid-1', 'ip_prefix':'10.0.1.1/24'})
        restapi_client.post_config_vlan(409, {'vnet_id' : 'vnet-guid-1', 'ip_prefix':'10.0.2.1/24'})

        # get vlans for vnet-guid-1
        r = restapi_client.get_config_interface_vlans('')
        j = json.loads(r.text)
        assert r.status_code ==404

    # Vlan Member
    def test_post_vlan_mem_which_exists_tagged(self, setup_restapi_client):
        _, _, _, restapi_client = setup_restapi_client
        restapi_client.post_generic_vlan_and_deps()
        r = restapi_client.post_config_vlan(3, {'vnet_id' : 'vnet-guid-1', 'ip_prefix' : '10.1.1.0/24'})
        assert r.status_code == 204
        attr = {'tagging_mode' : 'tagged'}
        r = restapi_client.post_config_vlan_member(2, "Ethernet1", attr)
        assert r.status_code == 204

        r = restapi_client.post_config_vlan_member(2, "Ethernet1", attr)
        assert r.status_code == 409
        j = json.loads(r.text)
        assert RESRC_EXISTS == j['error']['sub-code']

        r = restapi_client.post_config_vlan_member(3, "Ethernet1", attr)
        assert r.status_code == 204

        
    def test_post_vlan_mem_which_exists_untagged(self, setup_restapi_client):
        _, _, _, restapi_client = setup_restapi_client
        restapi_client.post_generic_vlan_and_deps()
        r = restapi_client.post_config_vlan(3, {'vnet_id' : 'vnet-guid-1', 'ip_prefix' : '10.1.1.0/24'})
        assert r.status_code == 204
        attr = {'tagging_mode' : 'untagged'}

        r = restapi_client.post_config_vlan_member(2, "Ethernet1", attr)
        assert r.status_code == 204

        r = restapi_client.post_config_vlan_member(2, "Ethernet1", attr)
        assert r.status_code == 409
        j = json.loads(r.text)
        assert RESRC_EXISTS == j['error']['sub-code']

        r = restapi_client.post_config_vlan_member(3, "Ethernet1", attr)
        assert r.status_code == 409
        j = json.loads(r.text)
        assert RESRC_EXISTS == j['error']['sub-code']

        r = restapi_client.delete_config_vlan_member(2, "Ethernet1")
        assert r.status_code == 204
        r = restapi_client.post_config_vlan_member(3, "Ethernet1", attr)
        assert r.status_code == 204

        attr = {'tagging_mode' : 'tagged'}
        r = restapi_client.post_config_vlan_member(2, "Ethernet1", attr)
        assert r.status_code == 204

    def test_get_vlan_member_not_created(self, setup_restapi_client):
        _, _, _, restapi_client = setup_restapi_client
        restapi_client.post_generic_vlan_and_deps()
        r = restapi_client.get_config_vlan_member(2, "ethernet2")
        assert r.status_code == 404
        j = json.loads(r.text)
        assert ['if_name'] == j['error']['fields']

    def test_delete_vlan_member_not_created(self, setup_restapi_client):
        _, _, _, restapi_client = setup_restapi_client
        restapi_client.post_generic_vlan_and_deps()
        r = restapi_client.delete_config_vlan_member(2, "ethernet2")
        assert r.status_code == 404
        j = json.loads(r.text)
        assert ['if_name'] == j['error']['fields']

    # Vlan Neighbor
    def test_post_vlan_neighbor_which_exists(self, setup_restapi_client):
        _, _, _, restapi_client = setup_restapi_client
        restapi_client.post_generic_vlan_and_deps()
        r = restapi_client.post_config_vlan_neighbor(2, "10.10.10.10")
        assert r.status_code == 204

        r = restapi_client.post_config_vlan_neighbor(2, "10.10.10.10")
        assert r.status_code == 409

        j = json.loads(r.text)
        assert RESRC_EXISTS == j['error']['sub-code']      

    def test_get_vlan_neighbor_not_created(self, setup_restapi_client):
        _, _, _, restapi_client = setup_restapi_client
        restapi_client.post_generic_vlan_and_deps()
        r = restapi_client.get_config_vlan_neighbor(2, "10.10.10.10")
        assert r.status_code == 404
        j = json.loads(r.text)
        assert ['ip_addr'] == j['error']['fields']

    def test_delete_vlan_neighbor_not_created(self, setup_restapi_client):
        _, _, _, restapi_client = setup_restapi_client
        restapi_client.post_generic_vlan_and_deps()
        r = restapi_client.delete_config_vlan_neighbor(2, "10.10.10.10")
        assert r.status_code == 404
        j = json.loads(r.text)
        assert ['ip_addr'] == j['error']['fields']

    def test_vlan_neighbor_not_valid_ip(self, setup_restapi_client):
        _, _, _, restapi_client = setup_restapi_client
        restapi_client.post_generic_vlan_and_deps()
        # post
        r = restapi_client.post_config_vlan_neighbor(2, "a.b.c.d")
        assert r.status_code == 400
        j = json.loads(r.text)
        assert ['ip_addr'] == j['error']['fields']

        # get
        r = restapi_client.get_config_vlan_neighbor(2, "a.b.c.d")
        assert r.status_code == 400
        j = json.loads(r.text)
        assert ['ip_addr'] == j['error']['fields']
 
        # delete
        r = restapi_client.delete_config_vlan_neighbor(2, "a.b.c.d")
        assert r.status_code == 400
        j = json.loads(r.text)
        assert ['ip_addr'] == j['error']['fields']

    # Routes
    def test_patch_delete_routes_not_created(self, setup_restapi_client):
        _, _, _, restapi_client = setup_restapi_client
        restapi_client.post_generic_vlan_and_deps()
        routes = []
        for i in range (1,100):
             routes.append({'cmd':'delete',
                            'ip_prefix':'10.2.'+str(i)+'.0/24',
                            'nexthop':'192.168.2.'+str(i),
                            'vnid': 1 + i%5,
                            'mac_address':'00:08:aa:bb:cd:'+hex(15+i)[2:]})

        # Patch
        r = restapi_client.patch_config_vrouter_vrf_id_routes("vnet-guid-1", routes)
        assert r.status_code == 207
        j = json.loads(r.text)
        for route in routes:
             route['error_code'] = 404
             route['error_msg'] = 'Not found'
        assert sorted(routes) == sorted(j['failed'])
        restapi_client.check_routes_dont_exist_in_tun_tb(1, routes)

    # Operations
    # PingVRF
    def test_post_ping_invalid(self, setup_restapi_client):
        _, _, _, restapi_client = setup_restapi_client
        restapi_client.post_generic_vrouter_and_deps()
        # Invalid count scenario
        r = restapi_client.post_ping({"count" : "abc", "ip_addr" : "8.8.8.8"})
        assert r.status_code == 400
        # Invalid ip_addr scenario
        r = restapi_client.post_ping({"ip_addr" : "8.8.8.888"})
        assert r.status_code == 400
        # vnet_id not found 404 error
        r = restapi_client.post_ping({'vnet_id' : 'vnet-1', 'ip_addr' : '8.8.8.8'})
        assert r.status_code == 404
