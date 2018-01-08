#!/usr/bin/env python3

import datetime
import json
import requests
import time
import unittest
import logging
import redis
import json

TEST_HOST = 'http://localhost:8080/'

logging.basicConfig(filename='test.log', filemode='w', level=logging.INFO)
l = logging.getLogger('msee_client')

class msee_client(unittest.TestCase):
    def put(self, url, body = []):
        if body == None:
            data = None
        else:
            data = json.dumps(body)

        l.info("Request PUT: %s" % url)
        l.info("JSON Body: %s" % data)
        r = requests.put(TEST_HOST + url, data=data, headers={'Content-Type': 'application/json'})
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

    # /config/vrouter/{vrf_id}
    def put_config_vrouter_vrf_id(self, vrf_id, value):
        return self.put('v1/config/vrouter/{vrf_id}'.format(vrf_id=vrf_id), value)

    def get_config_vrouter_vrf_id(self, vrf_id):
        return self.get('v1/config/vrouter/{vrf_id}'.format(vrf_id=vrf_id))

    def delete_config_vrouter_vrf_id(self, vrf_id):
        return self.delete('v1/config/vrouter/{vrf_id}'.format(vrf_id=vrf_id))

    # /config/vrouter
    def get_config_vrouter(self):
        return self.get('v1/config/vrouter')

    # /config/tunnel/encap/vxlan/{vnid}
    def put_config_tunnel_encap_vxlan_vnid(self, vnid, value):
        return self.put('v1/config/tunnel/encap/vxlan/{vnid}'.format(vnid=vnid), value)

    def delete_config_tunnel_encap_vxlan_vnid(self, vnid):
        return self.delete('v1/config/tunnel/encap/vxlan/{vnid}'.format(vnid=vnid))

    def get_config_tunnel_encap_vxlan_vnid(self, vnid):
        return self.get('v1/config/tunnel/encap/vxlan/{vnid}'.format(vnid=vnid))

    # /config/tunnel/encap/vxlan
    def get_config_tunnel_encap_vxlan(self):
        return self.get('v1/config/tunnel/encap/vxlan')

    # /config/vrouter/{vrf_id}/routes
    def put_config_vrouter_vrf_id_routes(self, vrf_id, value):
        return self.put('v1/config/vrouter/{vrf_id}/routes'.format(vrf_id=vrf_id), value)

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

    # /config/interface/qinq/{port}/{stag}/{ctag}
    def put_config_interface_qinq_port_stag_ctag(self, port, stag, ctag, value):
        return self.put('v1/config/interface/qinq/{port}/{stag}/{ctag}'.format(port=port, stag=stag, ctag=ctag), value)

    def delete_config_interface_qinq_port_stag_ctag(self, port, stag, ctag):
        return self.delete('v1/config/interface/qinq/{port}/{stag}/{ctag}'.format(port=port, stag=stag, ctag=ctag))

    def get_config_interface_qinq_port_stag_ctag(self, port, stag, ctag):
        return self.get('v1/config/interface/qinq/{port}/{stag}/{ctag}'.format(port=port, stag=stag, ctag=ctag))

    # /config/interface/qinq/{port}
    def delete_config_interface_qinq_port(self, port):
        return self.delete('v1/config/interface/qinq/{port}'.format(port=port))

    def get_config_interface_qinq_port(self, port):
        return self.get('v1/config/interface/qinq/{port}'.format(port=port))

    # /config/interface/port/{port}
    def put_config_interface_port_port(self, port, value):
        return self.put('v1/config/interface/port/{port}'.format(port=port), value)

    def delete_config_interface_port_port(self, port):
        return self.delete('v1/config/interface/port/{port}'.format(port=port))

    def get_config_interface_port_port(self, port):
        return self.get('v1/config/interface/port/{port}'.format(port=port))

    # /state/interface/{port}
    def get_state_interface_port(self, port):
        return self.get('v1/state/interface/{port}'.format(port=port))

    # /state/interface
    def get_state_interface(self):
        return self.get('v1/state/interface')

    # /config/tunnel/decap/tunnel_type
    def put_config_tunnel_decap_tunnel_type(self, tunnel_type, value):
        return self.put('v1/config/tunnel/decap/{tunnel_type}'.format(tunnel_type=tunnel_type), value)

    def get_config_tunnel_decap_tunnel_type(self, tunnel_type):
        return self.get('v1/config/tunnel/decap/{tunnel_type}'.format(tunnel_type=tunnel_type))

    def delete_config_tunnel_decap_tunnel_type(self, tunnel_type):
        return self.delete('v1/config/tunnel/decap/{tunnel_type}'.format(tunnel_type=tunnel_type))

    # Test setup
    def setUp(self):
        l.info('============================================================')
        l.info("Running: {0}".format(self._testMethodName))
        l.info('------------------------------------------------------------')

        # Clear DBs - reach known state
        self.db = redis.StrictRedis('localhost', 6379, 0)
        self.db.flushdb()

        self.cache = redis.StrictRedis('localhost', 6379, 4)
        self.cache.flushdb()

        # Sanity check
        keys = self.db.keys()
        self.assertEqual(keys, [])

        keys = self.cache.keys()
        self.assertEqual(keys, [])

    @classmethod
    def setUpClass(cls):
        l.info('============================================================')
        l.info("Starting: {0} - {1}".format(cls.__name__, cls.__doc__))
        l.info('------------------------------------------------------------')

class msee_expected_tests(msee_client):
    """Normal behaviour tests"""
    # /config/vrouter/{vrf_id}
    def test_put_config_vrouter_vrf_id(self):
        r = self.put_config_vrouter_vrf_id(1234, {
            'vrf_name': 'test_vrf_name'
        })
        self.assertEqual(r.status_code, 204)

        keys = self.db.keys()
        self.assertEqual(sorted(keys), sorted([b'VROUTER_TABLE:1234', b'VROUTER_TABLE_KEY_SET']))

        vrouter_table = self.db.hgetall('VROUTER_TABLE:1234')
        self.assertEqual(vrouter_table, {b'name': b'test_vrf_name'})

    def test_get_config_vrouter_vrf_id(self):
        self.db.hset('VROUTER_TABLE:1234', b'name', b'test_vrf_name')
        self.cache.hset('VRFNAME_VRFID_MAP', 'test_vrf_name', 1234)
        self.cache.hset('VRFID_VRFNAME_MAP', 1234, 'test_vrf_name')

        r = self.get_config_vrouter_vrf_id(1234)
        self.assertEqual(r.status_code, 200)

        j = json.loads(r.text)
        self.assertEqual(j, {
            'vrf_id': 1234,
            'attr': {
                'vrf_name': 'test_vrf_name'
            }
        })

    def test_delete_config_vrouter_vrf_id(self):
        self.db.hset('VROUTER_TABLE:1234', b'name', b'test_vrf_name')
        self.db.sadd('VROUTER_TABLE_KEY_SET', b'1234')
        self.cache.hset('VRFNAME_VRFID_MAP', 'test_vrf_name', 1234)
        self.cache.hset('VRFID_VRFNAME_MAP', 1234, 'test_vrf_name')

        r = self.delete_config_vrouter_vrf_id(1234)
        self.assertEqual(r.status_code, 204)

        keys = self.db.keys()
        self.assertEqual(sorted(keys), sorted([b'VROUTER_TABLE_KEY_SET']))

    # /config/vrouter
    def test_get_config_vrouter(self):
        self.db.hset('VROUTER_TABLE:1234', b'name', b'test_vrf_name')

        r = self.get_config_vrouter()
        self.assertEqual(r.status_code, 200)

        j = json.loads(r.text)
        self.assertEqual(j, [
            {
                'vrf_id': 1234,
                'attr': {
                    'vrf_name': 'test_vrf_name'
                }
            }
        ])

    def test_get_config_vrouter_multiple(self):
        self.db.hset('VROUTER_TABLE:1234', b'name', b'test_vrf_name1')
        self.db.hset('VROUTER_TABLE:1235', b'name', b'test_vrf_name2')
        self.db.hset('VROUTER_TABLE:1236', b'name', b'test_vrf_name3')
        self.db.hset('VROUTER_TABLE:1237', b'name', b'test_vrf_name4')

        r = self.get_config_vrouter()
        self.assertEqual(r.status_code, 200)

        j = json.loads(r.text)
        self.assertEqual(len(j), 4)

    # /config/tunnel/encap/vxlan/{vnid}
    def test_put_config_tunnel_encap_vxlan_vnid(self):
        self.cache.hset('VRFNAME_VRFID_MAP', 'testvrfname', 9876)
        self.cache.hset('VRFID_VRFNAME_MAP', 9876, 'testvrfname')

        r = self.put_config_tunnel_encap_vxlan_vnid(1234, {
            'vrf_id': 9876
        })
        self.assertEqual(r.status_code, 204)

        keys = self.db.keys()
        self.assertEqual(sorted(keys), sorted([b'TUNNEL_TABLE:encapsulation:vxlan:1234', b'TUNNEL_TABLE_KEY_SET']))

        tunnel_table = self.db.hgetall('TUNNEL_TABLE:encapsulation:vxlan:1234')
        self.assertEqual(tunnel_table, {b'vrf_id': b'9876'})

    def test_delete_config_tunnel_encap_vxlan_vnid(self):
        self.db.hset('TUNNEL_TABLE:encapsulation:vxlan:1234', b'vrf_id', b'9876')
        r = self.delete_config_tunnel_encap_vxlan_vnid(1234)
        self.assertEqual(r.status_code, 204)

        keys = self.db.keys()
        self.assertEqual(sorted(keys), sorted([b'TUNNEL_TABLE_KEY_SET']))

    def test_get_config_tunnel_encap_vxlan_vnid(self):
        self.db.hset('TUNNEL_TABLE:encapsulation:vxlan:1234', b'vrf_id', b'9876')
        r = self.get_config_tunnel_encap_vxlan_vnid(1234)
        self.assertEqual(r.status_code, 200)

        j = json.loads(r.text)
        self.assertEqual(j, {
            'vnid': 1234,
            'attr': {
                'vrf_id': 9876
            }
        })

    # /config/tunnel/encap/vxlan
    def test_get_config_tunnel_encap_vxlan(self):
        self.db.hset('TUNNEL_TABLE:encapsulation:vxlan:1234', b'vrf_id', b'9876')
        r = self.get_config_tunnel_encap_vxlan()
        self.assertEqual(r.status_code, 200)

        j = json.loads(r.text)
        self.assertEqual(j, [{
            'vnid': 1234,
            'attr': {
                'vrf_id': 9876
            }
        }])

    def test_get_config_tunnel_encap_vxlan_multiple(self):
        self.db.hset('TUNNEL_TABLE:encapsulation:vxlan:1234', b'vrf_id', b'9876')
        self.db.hset('TUNNEL_TABLE:encapsulation:vxlan:1235', b'vrf_id', b'9877')
        self.db.hset('TUNNEL_TABLE:encapsulation:vxlan:1236', b'vrf_id', b'9878')
        self.db.hset('TUNNEL_TABLE:encapsulation:vxlan:1237', b'vrf_id', b'9879')
        self.db.hset('TUNNEL_TABLE:encapsulation:vxlan:1238', b'vrf_id', b'9870')
        r = self.get_config_tunnel_encap_vxlan()
        self.assertEqual(r.status_code, 200)

        j = json.loads(r.text)
        self.assertEqual(5, len(j))

    # /config/vrouter/{vrf_id}/routes
    def test_put_config_vrouter_vrf_id_routes(self):
        self.cache.hset('VRFNAME_VRFID_MAP', 'testvrfname', 1234)
        self.cache.hset('VRFID_VRFNAME_MAP', 1234, 'testvrfname')
        self.db.hset('TUNNEL_TABLE:encapsulation:vxlan:9876', b'vrf_id', b'1234')

        r = self.put_config_vrouter_vrf_id_routes(1234, [
            {
                'ip_prefix': '127.0.0.0/24',
                'nexthop_type' : 'vxlan-tunnel',
                'nexthop': '92.89.1.2',
                'mac_address': '01:02:03:04:05:06',
                'vnid': 9876,
                'src_ip': '192.168.0.1',
                'error': 'none'
            }
        ])
        self.assertEqual(r.status_code, 201)

        keys = self.db.keys()
        self.assertEqual(sorted(keys), sorted([b'VROUTER_ROUTES_TABLE_KEY_SET', b'VROUTER_ROUTES_TABLE:1234:127.0.0.0/24', b'TUNNEL_TABLE:encapsulation:vxlan:9876']))

        vrouter_table = self.db.hgetall('VROUTER_ROUTES_TABLE:1234:127.0.0.0/24')
        self.assertEqual(vrouter_table, {
            b'nexthop': b'92.89.1.2',
            b'nexthop_type': b'vxlan-tunnel',
            b'mac_address': b'01:02:03:04:05:06',
            b'vxlanid': b'9876'
        })

    def test_put_config_vrouter_vrf_id_routes_standard(self):
        self.cache.hset('VRFNAME_VRFID_MAP', 'testvrfname', 1234)
        self.cache.hset('VRFID_VRFNAME_MAP', 1234, 'testvrfname')
        self.db.hset('TUNNEL_TABLE:encapsulation:vxlan:9876', b'vrf_id', b'1234')

        r = self.put_config_vrouter_vrf_id_routes(1234, [
            {
                'ip_prefix': '127.0.0.0/24',
                'nexthop_type' : 'vxlan-tunnel',
                'nexthop': '92.89.1.2',
                'mac_address': '01:02:03:04:05:06',
                'vnid': 9876,
                'src_ip': '192.168.0.1',
                'port': 'standard',
                'error': 'none'
            }
        ])
        self.assertEqual(r.status_code, 201)

        keys = self.db.keys()
        self.assertEqual(sorted(keys), sorted([b'VROUTER_ROUTES_TABLE_KEY_SET', b'VROUTER_ROUTES_TABLE:1234:127.0.0.0/24', b'TUNNEL_TABLE:encapsulation:vxlan:9876']))

        vrouter_table = self.db.hgetall('VROUTER_ROUTES_TABLE:1234:127.0.0.0/24')
        self.assertEqual(vrouter_table, {
            b'nexthop': b'92.89.1.2',
            b'nexthop_type': b'vxlan-tunnel',
            b'mac_address': b'01:02:03:04:05:06',
            b'vxlanid': b'9876'
        })

    @unittest.expectedFailure
    def test_put_config_vrouter_vrf_id_routes_ipv6(self):
        self.cache.hset('VRFNAME_VRFID_MAP', 'testvrfname', 1234)
        self.cache.hset('VRFID_VRFNAME_MAP', 1234, 'testvrfname')
        self.db.hset('TUNNEL_TABLE:encapsulation:vxlan:9876', b'vrf_id', b'1234')

        r = self.put_config_vrouter_vrf_id_routes(1234, [
            {
                'ip_prefix': '2620:0:ccc::2/32',
                'nexthop_type' : 'vxlan-tunnel',
                'nexthop': '2620:0:ccc::2',
                'mac_address': '01:02:03:04:05:06',
                'vnid': 9876,
                'src_ip': '192.168.0.1',
                'error': 'none'
            }
        ])
        self.assertEqual(r.status_code, 201)

        keys = self.db.keys()
        self.assertEqual(sorted(keys), sorted([b'VROUTER_ROUTES_TABLE_KEY_SET', b'VROUTER_ROUTES_TABLE:1234:2620:0:ccc::2/32', b'TUNNEL_TABLE:encapsulation:vxlan:9876']))

        vrouter_table = self.db.hgetall('VROUTER_ROUTES_TABLE:1234:2620:0:ccc::2/32')
        self.assertEqual(vrouter_table, {
            b'nexthop': b'2620:0:ccc::2',
            b'nexthop_type': b'vxlan-tunnel',
            b'mac_address': b'01:02:03:04:05:06',
            b'vxlanid': b'9876'
        })

    def test_put_config_vrouter_vrf_id_routes_nexthop_type_ip(self):
        self.cache.hset('VRFNAME_VRFID_MAP', 'testvrfname', 1234)
        self.cache.hset('VRFID_VRFNAME_MAP', 1234, 'testvrfname')
        self.cache.hset('VRFID_VLANID_MAP', 1234, 2)
        self.db.hmset('QINQ_TABLE:Ethernet16:5:6', {
            b'vrf_id': b'1234',
            b'peer_ip': b'5.6.7.8',
            b'proxy_arp_ip': b'1.2.3.4',
            b'subnet': b'1.2.3.0/24'
        })

        r = self.put_config_vrouter_vrf_id_routes(1234, [
            {
                'ip_prefix': '127.0.0.0/24',
                'nexthop_type' : 'ip',
                'nexthop': '12.34.56.78',
                'src_ip': '192.168.0.1',
                'error': 'none'
            }
        ])
        self.assertEqual(r.status_code, 201)

        keys = self.db.keys()
        self.assertEqual(sorted(keys), sorted([b'VROUTER_ROUTES_TABLE_KEY_SET', b'VROUTER_ROUTES_TABLE:1234:127.0.0.0/24',b'QINQ_TABLE:Ethernet16:5:6']))

        vrouter_table = self.db.hgetall('VROUTER_ROUTES_TABLE:1234:127.0.0.0/24')
        self.assertEqual(vrouter_table, {
            b'nexthop': b'12.34.56.78',
            b'nexthop_type': b'ip',
            b'mac_address': b'12:34:56:78:90:12'
        })

    def test_put_config_vrouter_vrf_id_routes_nexthop_type_ip_multi(self):
        self.cache.hset('VRFNAME_VRFID_MAP', 'testvrfname', 1234)
        self.cache.hset('VRFID_VRFNAME_MAP', 1234, 'testvrfname')
        self.cache.hset('VRFID_VLANID_MAP', 1234, 2)
        self.db.hmset('QINQ_TABLE:Ethernet16:5:6', {
            b'vrf_id': b'1234',
            b'peer_ip': b'5.6.7.8',
            b'proxy_arp_ip': b'1.2.3.4',
            b'subnet': b'1.2.3.0/24'
        })

        r = self.put_config_vrouter_vrf_id_routes(1234, [
            {
                'ip_prefix': '127.0.0.0/24',
                'nexthop_type' : 'ip',
                'nexthop': '12.34.56.78',
                'src_ip': '192.168.0.1',
                'error': 'none'
            },
            {
                'ip_prefix': '127.1.0.0/24',
                'nexthop_type' : 'ip',
                'nexthop': '92.89.1.3',
                'src_ip': '192.168.0.1',
                'error': 'none'
            },
            {
                'ip_prefix': '127.2.0.0/24',
                'nexthop_type' : 'ip',
                'nexthop': '90.12.34.56',
                'src_ip': '192.168.0.1',
                'error': 'none'
            }
        ])
        self.assertEqual(r.status_code, 201)

        keys = self.db.keys()
        self.assertEqual(sorted(keys), sorted([
            b'VROUTER_ROUTES_TABLE_KEY_SET',
            b'VROUTER_ROUTES_TABLE:1234:127.0.0.0/24',
            b'VROUTER_ROUTES_TABLE:1234:127.2.0.0/24',
            b'QINQ_TABLE:Ethernet16:5:6'
        ]))

        vrouter_table = self.db.hgetall('VROUTER_ROUTES_TABLE:1234:127.0.0.0/24')
        self.assertEqual(vrouter_table, {
            b'nexthop': b'12.34.56.78',
            b'nexthop_type': b'ip',
            b'mac_address': b'12:34:56:78:90:12'
        })

        vrouter_table = self.db.hgetall('VROUTER_ROUTES_TABLE:1234:127.2.0.0/24')
        self.assertEqual(vrouter_table, {
            b'nexthop': b'90.12.34.56',
            b'nexthop_type': b'ip',
            b'mac_address': b'34:56:78:90:12:34'
        })

    def test_put_config_vrouter_vrf_id_routes_src_error_missing(self):
        self.cache.hset('VRFNAME_VRFID_MAP', 'testvrfname', 1234)
        self.cache.hset('VRFID_VRFNAME_MAP', 1234, 'testvrfname')
        self.db.hset('TUNNEL_TABLE:encapsulation:vxlan:9876', b'vrf_id', b'1234')

        r = self.put_config_vrouter_vrf_id_routes(1234, [
            {
                'ip_prefix': '127.0.0.0/24',
                'nexthop_type' : 'vxlan-tunnel',
                'nexthop': '92.89.1.2',
                'mac_address': '01:02:03:04:05:06',
                'vnid': 9876,
                'src_ip': '192.168.0.1',
            }
        ])
        self.assertEqual(r.status_code, 201)

        keys = self.db.keys()
        self.assertEqual(sorted(keys), sorted([b'VROUTER_ROUTES_TABLE_KEY_SET', b'VROUTER_ROUTES_TABLE:1234:127.0.0.0/24', b'TUNNEL_TABLE:encapsulation:vxlan:9876']))

        vrouter_table = self.db.hgetall('VROUTER_ROUTES_TABLE:1234:127.0.0.0/24')
        self.assertEqual(vrouter_table, {
            b'nexthop': b'92.89.1.2',
            b'nexthop_type': b'vxlan-tunnel',
            b'mac_address': b'01:02:03:04:05:06',
            b'vxlanid': b'9876'
        })

    def test_delete_config_vrouter_vrf_id_routes(self):
        self.cache.hset('VRFNAME_VRFID_MAP', 'testvrfname', 1234)
        self.cache.hset('VRFID_VRFNAME_MAP', 1234, 'testvrfname')
        self.db.hmset('VROUTER_ROUTES_TABLE:1234:127.0.0.0/24', {
            b'nexthop_type': b'vxlan-tunnel',
            b'nexthop': b'92.89.1.2',
            b'mac_address': b'01:02:03:04:05:06',
            b'vxlanid': 9876
        })
        self.db.sadd('VROUTER_ROUTES_TABLE_KEY_SET', b'1234:127.0.0.0/24')

        r = self.delete_config_vrouter_vrf_id_routes(1234, 9876, [
            {
                'ip_prefix': '127.0.0.0/24',
                'nexthop_type' : 'vxlan-tunnel',
                'nexthop': '92.89.1.2',
                'mac_address': '01:02:03:04:05:06',
                'vnid': 9876,
                'src_ip': '192.168.0.1',
                'error': 'none'
            }
        ])
        self.assertEqual(r.status_code, 201)

        j = json.loads(r.text)
        self.assertEqual(j, {
            'removed': [{
                'ip_prefix': '127.0.0.0/24',
                'nexthop_type' : 'vxlan-tunnel',
                'nexthop': '92.89.1.2',
                'mac_address': '01:02:03:04:05:06',
                'vnid': 9876,
                'src_ip': '192.168.0.1',
                'port': 'azure',
                'error': 'none'
            }]
        })

        keys = self.db.keys()
        self.assertEqual(sorted(keys), sorted([b'VROUTER_ROUTES_TABLE_KEY_SET']))

    def test_delete_config_vrouter_vrf_id_routes_all(self):
        self.cache.hset('VRFNAME_VRFID_MAP', 'testvrfname', 1234)
        self.cache.hset('VRFID_VRFNAME_MAP', 1234, 'testvrfname')
        self.db.hmset('VROUTER_ROUTES_TABLE:1234:127.0.0.0/24', {
            b'nexthop_type': b'vxlan-tunnel',
            b'nexthop': b'92.89.1.2',
            b'mac_address': b'01:02:03:04:05:06',
            b'vxlanid': 9876
        })
        self.db.sadd('VROUTER_ROUTES_TABLE_KEY_SET', b'1234:127.0.0.0/24')

        r = self.delete_config_vrouter_vrf_id_routes(1234)
        self.assertEqual(r.status_code, 201)

        j = json.loads(r.text)
        self.assertEqual(j, {
            'removed': [{
                'ip_prefix': '127.0.0.0/24',
                'nexthop_type' : 'vxlan-tunnel',
                'nexthop': '92.89.1.2',
                'mac_address': '01:02:03:04:05:06',
                'vnid': 9876
            }]
        })

        keys = self.db.keys()
        self.assertEqual(sorted(keys), sorted([b'VROUTER_ROUTES_TABLE_KEY_SET']))

    def test_delete_config_vrouter_vrf_id_routes_filter(self):
        self.cache.hset('VRFNAME_VRFID_MAP', 'testvrfname', 1234)
        self.cache.hset('VRFID_VRFNAME_MAP', 1234, 'testvrfname')
        self.db.hmset('VROUTER_ROUTES_TABLE:1234:127.0.0.0/24', {
            b'nexthop_type': b'vxlan-tunnel',
            b'nexthop': b'92.89.1.2',
            b'mac_address': b'01:02:03:04:05:06',
            b'vxlanid': 9876
        })
        self.db.hmset('VROUTER_ROUTES_TABLE:1234:127.0.0.1/24', {
            b'nexthop_type': b'vxlan-tunnel',
            b'nexthop': b'92.89.1.2',
            b'mac_address': b'01:02:03:04:05:06',
            b'vxlanid': 9875
        })
        self.db.sadd('VROUTER_ROUTES_TABLE_KEY_SET', b'1234:127.0.0.0/24', b'1234:127.0.0.1/24')

        r = self.delete_config_vrouter_vrf_id_routes(1234, 9876)
        self.assertEqual(r.status_code, 201)

        j = json.loads(r.text)
        self.assertEqual(j, {
            'removed': [{
                'ip_prefix': '127.0.0.0/24',
                'nexthop_type' : 'vxlan-tunnel',
                'nexthop': '92.89.1.2',
                'mac_address': '01:02:03:04:05:06',
                'vnid': 9876
            }]
        })

        keys = self.db.keys()
        self.assertEqual(sorted(keys), sorted([b'VROUTER_ROUTES_TABLE_KEY_SET', b'VROUTER_ROUTES_TABLE:1234:127.0.0.1/24']))

    def test_delete_config_vrouter_vrf_id_routes_novnid(self):
        self.cache.hset('VRFNAME_VRFID_MAP', 'testvrfname', 1234)
        self.cache.hset('VRFID_VRFNAME_MAP', 1234, 'testvrfname')
        self.db.hmset('VROUTER_ROUTES_TABLE:1234:127.0.0.0/24', {
            b'nexthop_type': b'vxlan-tunnel',
            b'nexthop': b'92.89.1.2',
            b'mac_address': b'01:02:03:04:05:06',
            b'vxlanid': 9876
        })
        self.db.sadd('VROUTER_ROUTES_TABLE_KEY_SET', b'1234:127.0.0.0/24')

        r = self.delete_config_vrouter_vrf_id_routes(1234, None, [
            {
                'ip_prefix': '127.0.0.0/24',
                'nexthop_type' : 'vxlan-tunnel',
                'nexthop': '92.89.1.2',
                'mac_address': '01:02:03:04:05:06',
                'vnid': 9876,
                'src_ip': '192.168.0.1',
                'error': 'none'
            }
        ])
        self.assertEqual(r.status_code, 201)

        j = json.loads(r.text)
        self.assertEqual(j, {
            'removed': [{
                'ip_prefix': '127.0.0.0/24',
                'nexthop_type' : 'vxlan-tunnel',
                'nexthop': '92.89.1.2',
                'mac_address': '01:02:03:04:05:06',
                'vnid': 9876,
                'src_ip': '192.168.0.1',
                'port': 'azure',
                'error': 'none'
            }]
        })

        keys = self.db.keys()
        self.assertEqual(sorted(keys), sorted([b'VROUTER_ROUTES_TABLE_KEY_SET']))


    def test_delete_config_vrouter_vrf_id_routes_fail(self):
        self.cache.hset('VRFNAME_VRFID_MAP', 'testvrfname', 1234)
        self.cache.hset('VRFID_VRFNAME_MAP', 1234, 'testvrfname')
        r = self.delete_config_vrouter_vrf_id_routes(1234, 9876, [
            {
                'ip_prefix': '127.0.0.0/24',
                'nexthop_type' : 'vxlan-tunnel',
                'nexthop': '92.89.1.2',
                'mac_address': '01:02:03:04:05:06',
                'vnid': 9876,
                'src_ip': '192.168.0.1',
                'error': 'none'
            }
        ])
        self.assertEqual(r.status_code, 201)

        j = json.loads(r.text)
        self.assertEqual(j, {
            'failed': [{
                'ip_prefix': '127.0.0.0/24',
                'nexthop_type' : 'vxlan-tunnel',
                'nexthop': '92.89.1.2',
                'mac_address': '01:02:03:04:05:06',
                'vnid': 9876,
                'src_ip': '192.168.0.1',
                'port': 'azure',
                'error': 'none'
            }]
        })

    def test_delete_config_vrouter_vrf_id_routes_both(self):
        self.cache.hset('VRFNAME_VRFID_MAP', 'testvrfname', 1234)
        self.cache.hset('VRFID_VRFNAME_MAP', 1234, 'testvrfname')
        self.db.hmset('VROUTER_ROUTES_TABLE:1234:127.0.0.0/24', {
            b'nexthop_type': b'vxlan-tunnel',
            b'nexthop': b'92.89.1.2',
            b'mac_address': b'01:02:03:04:05:06',
            b'vxlanid': 9876
        })
        self.db.sadd('VROUTER_ROUTES_TABLE_KEY_SET', b'1234:127.0.0.0/24')

        r = self.delete_config_vrouter_vrf_id_routes(1234, 9876, [
            {
                'ip_prefix': '127.0.0.0/24',
                'nexthop_type' : 'vxlan-tunnel',
                'nexthop': '92.89.1.2',
                'mac_address': '01:02:03:04:05:06',
                'vnid': 9876,
                'src_ip': '192.168.0.1',
                'error': 'none'
            },
            {
                'ip_prefix': '92.89.1.2/24',
                'nexthop_type' : 'vxlan-tunnel',
                'nexthop': '192.168.0.1',
                'mac_address': '01:02:03:04:05:06',
                'vnid': 9876,
                'src_ip': '127.0.0.0',
                'error': 'none'
            }
        ])
        self.assertEqual(r.status_code, 201)

        j = json.loads(r.text)
        self.assertEqual(j, {
            'removed': [{
                'ip_prefix': '127.0.0.0/24',
                'nexthop_type' : 'vxlan-tunnel',
                'nexthop': '92.89.1.2',
                'mac_address': '01:02:03:04:05:06',
                'vnid': 9876,
                'src_ip': '192.168.0.1',
                'port': 'azure',
                'error': 'none'
            }],
            'failed': [{
                'ip_prefix': '92.89.1.2/24',
                'nexthop_type' : 'vxlan-tunnel',
                'nexthop': '192.168.0.1',
                'mac_address': '01:02:03:04:05:06',
                'vnid': 9876,
                'src_ip': '127.0.0.0',
                'port': 'azure',
                'error': 'none'
            }]
        })

        keys = self.db.keys()
        self.assertEqual(sorted(keys), sorted([b'VROUTER_ROUTES_TABLE_KEY_SET']))

    def test_get_config_vrouter_vrf_id_routes(self):
        self.cache.hset('VRFNAME_VRFID_MAP', 'testvrfname', 1234)
        self.cache.hset('VRFID_VRFNAME_MAP', 1234, 'testvrfname')
        self.db.hmset('VROUTER_ROUTES_TABLE:1234:127.0.0.0/24', {
            b'nexthop_type': b'vxlan-tunnel',
            b'nexthop': b'92.89.1.2',
            b'mac_address': b'01:02:03:04:05:06',
            b'vxlanid': 9876
        })

        self.db.hmset('VROUTER_ROUTES_TABLE:1254:127.0.0.9/24', {
            b'nexthop_type': b'vxlan-tunnel',
            b'nexthop': b'92.89.1.2',
            b'mac_address': b'01:02:03:04:05:06',
            b'vxlanid': 9876
        })

        self.db.hmset('VROUTER_ROUTES_TABLE:123:8.8.8.8/24', {
            b'nexthop_type': b'vxlan-tunnel',
            b'nexthop': b'92.89.1.2',
            b'mac_address': b'01:02:03:04:05:06',
            b'vxlanid': 9876
        })

        r = self.get_config_vrouter_vrf_id_routes(1234)
        self.assertEqual(r.status_code, 200)

        j = json.loads(r.text)
        self.assertEqual(j, [
            {
                'ip_prefix': '127.0.0.0/24',
                'nexthop_type': 'vxlan-tunnel',
                'nexthop': '92.89.1.2',
                'mac_address': '01:02:03:04:05:06',
                'vnid': 9876
            }
        ])

    def test_get_config_vrouter_vrf_id_routes_vnid_filter(self):
        self.cache.hset('VRFNAME_VRFID_MAP', 'testvrfname', 1234)
        self.cache.hset('VRFID_VRFNAME_MAP', 1234, 'testvrfname')
        self.db.hmset('VROUTER_ROUTES_TABLE:1234:127.0.0.0/24', {
            b'nexthop_type': b'vxlan-tunnel',
            b'nexthop': b'92.89.1.2',
            b'mac_address': b'01:02:03:04:05:06',
            b'vxlanid': 9856
        })

        self.db.hmset('VROUTER_ROUTES_TABLE:1234:127.0.0.9/24', {
            b'nexthop_type': b'vxlan-tunnel',
            b'nexthop': b'92.89.1.2',
            b'mac_address': b'01:02:03:04:05:06',
            b'vxlanid': 9876
        })

        self.db.hmset('VROUTER_ROUTES_TABLE:123:8.8.8.8/24', {
            b'nexthop_type': b'vxlan-tunnel',
            b'nexthop': b'92.89.1.2',
            b'mac_address': b'01:02:03:04:05:06',
            b'vxlanid': 9876
        })

        r = self.get_config_vrouter_vrf_id_routes(1234, 9876)
        self.assertEqual(r.status_code, 200)

        j = json.loads(r.text)
        self.assertEqual(j, [
            {
                'ip_prefix': '127.0.0.9/24',
                'nexthop_type': 'vxlan-tunnel',
                'nexthop': '92.89.1.2',
                'mac_address': '01:02:03:04:05:06',
                'vnid': 9876
            }
        ])

    def test_get_config_vrouter_vrf_id_routes_ip_filter(self):
        self.cache.hset('VRFNAME_VRFID_MAP', 'testvrfname', 1234)
        self.cache.hset('VRFID_VRFNAME_MAP', 1234, 'testvrfname')
        self.db.hmset('VROUTER_ROUTES_TABLE:1234:127.0.0.0/24', {
            b'nexthop_type': b'vxlan-tunnel',
            b'nexthop': b'92.89.1.2',
            b'mac_address': b'01:02:03:04:05:06',
            b'vxlanid': 9876
        })

        self.db.hmset('VROUTER_ROUTES_TABLE:1234:127.0.0.9/24', {
            b'nexthop_type': b'vxlan-tunnel',
            b'nexthop': b'92.89.1.2',
            b'mac_address': b'01:02:03:04:05:06',
            b'vxlanid': 9876
        })

        self.db.hmset('VROUTER_ROUTES_TABLE:123:8.8.8.8/24', {
            b'nexthop_type': b'vxlan-tunnel',
            b'nexthop': b'92.89.1.2',
            b'mac_address': b'01:02:03:04:05:06',
            b'vxlanid': 9876
        })

        r = self.get_config_vrouter_vrf_id_routes(1234, ip_prefix='127.0.0.9/24')
        self.assertEqual(r.status_code, 200)

        j = json.loads(r.text)
        self.assertEqual(j, [
            {
                'ip_prefix': '127.0.0.9/24',
                'nexthop_type': 'vxlan-tunnel',
                'nexthop': '92.89.1.2',
                'mac_address': '01:02:03:04:05:06',
                'vnid': 9876
            }
        ])

    # /config/interface/qinq/{port}/{stag}/{ctag}
    def test_put_config_interface_qinq_port_stag_ctag(self):
        r = self.put_config_interface_qinq_port_stag_ctag('Ethernet12', 5, 6, {
            'vrf_id': 1234,
            'peer_ip': '5.6.7.8',
            'proxy_arp_ip': '1.2.3.4',
            'subnet': '1.2.3.0/24'
        })
        self.assertEqual(r.status_code, 204)

        keys = self.db.keys()
        self.assertEqual(sorted(keys), sorted([
            b'QINQ_TABLE_KEY_SET',
            b'QINQ_TABLE:Ethernet12:5:6'
        ]))

        qinq_table = self.db.hgetall('QINQ_TABLE:Ethernet12:5:6')
        self.assertEqual(qinq_table, {
            b'vrf_id': b'1234',
            b'peer_ip': b'5.6.7.8',
            b'proxy_arp_ip': b'1.2.3.4',
            b'subnet': b'1.2.3.0/24'
        })

    def test_put_config_interface_qinq_port_stag_ctag_update(self):
        r = self.put_config_interface_qinq_port_stag_ctag('Ethernet12', 5, 6, {
            'vrf_id': 1234,
            'peer_ip': '5.6.7.8',
            'proxy_arp_ip': '1.2.3.4',
            'subnet': '1.2.3.0/24'
        })
        self.assertEqual(r.status_code, 204)

        r = self.put_config_interface_qinq_port_stag_ctag('Ethernet12', 5, 6, {
            'vrf_id': 1234,
            'peer_ip': '5.6.7.8',
            'proxy_arp_ip': '1.3.5.7',
            'subnet': '1.2.3.0/24'
        })
        self.assertEqual(r.status_code, 204)

        keys = self.db.keys()
        self.assertEqual(sorted(keys), sorted([
            b'QINQ_TABLE_KEY_SET',
            b'QINQ_TABLE:Ethernet12:5:6'
        ]))

        qinq_table = self.db.hgetall('QINQ_TABLE:Ethernet12:5:6')
        self.assertEqual(qinq_table, {
            b'vrf_id': b'1234',
            b'peer_ip': b'5.6.7.8',
            b'proxy_arp_ip': b'1.3.5.7',
            b'subnet': b'1.2.3.0/24'
        })

    def test_put_config_interface_qinq_port_stag_ctag_update_fail(self):
        r = self.put_config_interface_qinq_port_stag_ctag('Ethernet12', 5, 6, {
            'vrf_id': 1234,
            'peer_ip': '5.6.7.8',
            'proxy_arp_ip': '1.2.3.4',
            'subnet': '1.2.3.0/24'
        })
        self.assertEqual(r.status_code, 204)

        r = self.put_config_interface_qinq_port_stag_ctag('Ethernet12', 5, 6, {
            'vrf_id': 4321,
            'peer_ip': '5.6.7.8',
            'proxy_arp_ip': '1.3.5.7',
            'subnet': '1.2.3.0/24'
        })
        self.assertEqual(r.status_code, 405)

    def test_delete_config_interface_qinq_port_stag_ctag(self):
        self.db.hmset('QINQ_TABLE:Ethernet12:5:6', {
            b'vrf_id': b'1234',
            b'peer_ip': b'5.6.7.8',
            b'proxy_arp_ip': b'1.2.3.4',
            b'subnet': b'1.2.3.0/24'
        })
        self.db.sadd('QINQ_TABLE_KEY_SET', b'Ethernet12:5:6')

        r = self.delete_config_interface_qinq_port_stag_ctag('Ethernet12', 5, 6)
        self.assertEqual(r.status_code, 204)

        keys = self.db.keys()
        self.assertEqual(keys, [b'QINQ_TABLE_KEY_SET'])

        tag_set = self.cache.smembers('VRF_QINQ:1234')
        self.assertEqual(tag_set, set())

    def test_get_config_interface_qinq_port_stag_ctag(self):
        self.db.hmset('QINQ_TABLE:Ethernet0:5:6', {
            b'vrf_id': b'1234',
            b'peer_ip': b'5.6.7.8',
            b'proxy_arp_ip': b'1.2.3.4',
            b'subnet': b'1.2.3.0/24'
        })
        self.db.sadd('QINQ_TABLE_KEY_SET', b'Ethernet0:5:6')

        r = self.get_config_interface_qinq_port_stag_ctag('Ethernet0', 5, 6)
        self.assertEqual(r.status_code, 200)

        j = json.loads(r.text)
        self.assertEqual(j, {
            'port': 'Ethernet0',
            'stag': 5,
            'ctag': 6,
            'attr': {
                'vrf_id': 1234,
                'peer_ip': '5.6.7.8',
                'proxy_arp_ip': '1.2.3.4',
                'subnet': '1.2.3.0/24'
            }
        })

    # /config/interface/qinq/{port}
    def test_delete_config_interface_qinq_port(self):
        self.db.hmset('QINQ_TABLE:Ethernet12:5:6', {
            b'vrf_id': b'1234',
            b'peer_ip': b'5.6.7.8',
            b'proxy_arp_ip': b'1.2.3.4',
            b'subnet': b'1.2.3.0/24'
        })
        self.db.hmset('QINQ_TABLE:Ethernet16:5:6', {
            b'vrf_id': b'1234',
            b'peer_ip': b'5.6.7.8',
            b'proxy_arp_ip': b'1.2.3.4',
            b'subnet': b'1.2.3.0/24'
        })
        self.db.sadd('QINQ_TABLE_KEY_SET', b'Ethernet12:5:6', b'Ethernet16:5:6')

        r = self.delete_config_interface_qinq_port('Ethernet12')
        self.assertEqual(r.status_code, 204)

        keys = self.db.keys()
        self.assertEqual(sorted(keys), sorted([b'QINQ_TABLE_KEY_SET', b'QINQ_TABLE:Ethernet16:5:6']))

    def test_get_config_interface_qinq_port(self):
        self.db.hmset('QINQ_TABLE:Ethernet0:5:6', {
            b'vrf_id': b'1234',
            b'peer_ip': b'5.6.7.8',
            b'proxy_arp_ip': b'1.2.3.4',
            b'subnet': b'1.2.3.0/24'
        })
        self.db.sadd('QINQ_TABLE_KEY_SET', b'Ethernet0:5:6')

        r = self.get_config_interface_qinq_port('Ethernet0')
        self.assertEqual(r.status_code, 200)

        j = json.loads(r.text)
        self.assertEqual(j, [{
            'port': 'Ethernet0',
            'stag': 5,
            'ctag': 6,
            'attr': {
                'vrf_id': 1234,
                'peer_ip': '5.6.7.8',
                'proxy_arp_ip': '1.2.3.4',
                'subnet': '1.2.3.0/24'
            }
        }])

    # /config/interface/port/{port}
    def test_put_config_interface_port_port(self):
        self.cache.hset('VRFNAME_VRFID_MAP', 'testvrfname', 1234)
        self.cache.hset('VRFID_VRFNAME_MAP', 1234, 'testvrfname')

        r = self.put_config_interface_port_port('Ethernet12', {
            'vrf_id': 1234,
            'addr': '127.0.0.1/24',
            'mac_address': '01:02:03:04:05:06'
        })
        self.assertEqual(r.status_code, 204)

        keys = self.db.keys()
        self.assertEqual(sorted(keys), sorted([
            b'IGNORE_INTF_TABLE_KEY_SET',
            b'IGNORE_INTF_TABLE:Ethernet12:127.0.0.1/24',
            b'VLAN_MEMBER_TABLE:Vlan3840:Ethernet0',
            b'VLAN_MEMBER_TABLE_KEY_SET'
        ]))

        intf_table = self.db.hgetall('IGNORE_INTF_TABLE:Ethernet12:127.0.0.1/24')
        self.assertEqual(intf_table, {
            b'vrf_id': b'1234'
        })

        vlan_table = self.db.hgetall('VLAN_MEMBER_TABLE:Vlan3840:Ethernet0')
        self.assertEqual(vlan_table, {
            b'tagging_mode': b'tagged'
        })

        port_addr = self.cache.hget('PORT_ADDR_MAP', 'Ethernet12')
        self.assertEqual(b'127.0.0.1/24', port_addr)

        mac_addr = self.cache.hget('PORT_MAC_MAP', 'Ethernet12')
        self.assertEqual(b'01:02:03:04:05:06', mac_addr)

    def test_put_config_interface_port_port_spoof_guard(self):
        self.cache.hset('VRFNAME_VRFID_MAP', 'testvrfname', 1234)
        self.cache.hset('VRFID_VRFNAME_MAP', 1234, 'testvrfname')

        r = self.put_config_interface_port_port('Ethernet12', {
            'vrf_id': 1234,
            'addr': '127.0.0.1/24',
            'spoof_guard': ['127.0.0.1/24', '1.2.3.4/12', '5.4.3.2/18'],
            'mac_address': '01:02:03:04:05:06'
        })
        self.assertEqual(r.status_code, 204)

        keys = self.db.keys()
        self.assertEqual(sorted(keys), sorted([
            b'IGNORE_INTF_TABLE_KEY_SET',
            b'IGNORE_INTF_TABLE:Ethernet12:127.0.0.1/24',
            b'SPOOF_GUARD_TABLE_KEY_SET',
            b'SPOOF_GUARD_TABLE:Ethernet12',
            b'VLAN_MEMBER_TABLE:Vlan3840:Ethernet0',
            b'VLAN_MEMBER_TABLE_KEY_SET'
        ]))

        intf_table = self.db.hgetall('IGNORE_INTF_TABLE:Ethernet12:127.0.0.1/24')
        self.assertEqual(intf_table, {
            b'vrf_id': b'1234'
        })

        vlan_table = self.db.hgetall('VLAN_MEMBER_TABLE:Vlan3840:Ethernet0')
        self.assertEqual(vlan_table, {
            b'tagging_mode': b'tagged'
        })

        spoof_guard_table = self.db.hgetall('SPOOF_GUARD_TABLE:Ethernet12')
        self.assertEqual(spoof_guard_table, {
            b'addr_list': b'127.0.0.1/24,1.2.3.4/12,5.4.3.2/18'
        })

        port_addr = self.cache.hget('PORT_ADDR_MAP', 'Ethernet12')
        self.assertEqual(b'127.0.0.1/24', port_addr)

        mac_addr = self.cache.hget('PORT_MAC_MAP', 'Ethernet12')
        self.assertEqual(b'01:02:03:04:05:06', mac_addr)

    def test_put_config_interface_port_port_existing_port_ip_change(self):
        self.cache.hset('VRFNAME_VRFID_MAP', 'testvrfname', 1234)
        self.cache.hset('VRFID_VRFNAME_MAP', 1234, 'testvrfname')

        r = self.put_config_interface_port_port('Ethernet12', {
            'vrf_id': 1234,
            'addr': '127.0.0.0/24',
            'mac_address': '01:02:03:04:05:06'
        })
        self.assertEqual(r.status_code, 204)

        r = self.put_config_interface_port_port('Ethernet12', {
            'vrf_id': 1234,
            'addr': '127.0.0.1/24',
            'mac_address': '01:02:03:04:05:06'
        })
        self.assertEqual(r.status_code, 405)

        r = self.put_config_interface_port_port('Ethernet12', {
            'vrf_id': 1234,
            'addr': '127.0.0.0/32',
            'mac_address': '01:02:03:04:05:06'
        })
        self.assertEqual(r.status_code, 405)

        keys = self.db.keys()
        self.assertEqual(sorted(keys), sorted([
            b'IGNORE_INTF_TABLE_KEY_SET',
            b'IGNORE_INTF_TABLE:Ethernet12:127.0.0.0/24',
            b'VLAN_MEMBER_TABLE:Vlan3840:Ethernet0',
            b'VLAN_MEMBER_TABLE_KEY_SET'
        ]))

        intf_table = self.db.hgetall('IGNORE_INTF_TABLE:Ethernet12:127.0.0.0/24')
        self.assertEqual(intf_table, {
            b'vrf_id': b'1234'
        })

        vlan_table = self.db.hgetall('VLAN_MEMBER_TABLE:Vlan3840:Ethernet0')
        self.assertEqual(vlan_table, {
            b'tagging_mode': b'tagged'
        })

        port_addr = self.cache.hget('PORT_ADDR_MAP', 'Ethernet12')
        self.assertEqual(b'127.0.0.0/24', port_addr)

        mac_addr = self.cache.hget('PORT_MAC_MAP', 'Ethernet12')
        self.assertEqual(b'01:02:03:04:05:06', mac_addr)

    def test_put_config_interface_port_port_existing_port_mac_change(self):
        self.cache.hset('VRFNAME_VRFID_MAP', 'testvrfname', 1234)
        self.cache.hset('VRFID_VRFNAME_MAP', 1234, 'testvrfname')

        r = self.put_config_interface_port_port('Ethernet12', {
            'vrf_id': 1234,
            'addr': '127.0.0.0/24',
            'mac_address': '01:02:03:04:05:06'
        })
        self.assertEqual(r.status_code, 204)

        r = self.put_config_interface_port_port('Ethernet12', {
            'vrf_id': 1234,
            'addr': '127.0.0.0/24',
            'mac_address': '02:02:03:04:05:06'
        })
        self.assertEqual(r.status_code, 405)

        keys = self.db.keys()
        self.assertEqual(sorted(keys), sorted([
            b'IGNORE_INTF_TABLE_KEY_SET',
            b'IGNORE_INTF_TABLE:Ethernet12:127.0.0.0/24',
            b'VLAN_MEMBER_TABLE:Vlan3840:Ethernet0',
            b'VLAN_MEMBER_TABLE_KEY_SET'
        ]))

        intf_table = self.db.hgetall('IGNORE_INTF_TABLE:Ethernet12:127.0.0.0/24')
        self.assertEqual(intf_table, {
            b'vrf_id': b'1234'
        })

        vlan_table = self.db.hgetall('VLAN_MEMBER_TABLE:Vlan3840:Ethernet0')
        self.assertEqual(vlan_table, {
            b'tagging_mode': b'tagged'
        })

        port_addr = self.cache.hget('PORT_ADDR_MAP', 'Ethernet12')
        self.assertEqual(b'127.0.0.0/24', port_addr)

        mac_addr = self.cache.hget('PORT_MAC_MAP', 'Ethernet12')
        self.assertEqual(b'01:02:03:04:05:06', mac_addr)

    def test_put_config_interface_port_port_existing_port_vrf_change(self):
        self.cache.hset('VRFNAME_VRFID_MAP', 'testvrfname', 1234)
        self.cache.hset('VRFID_VRFNAME_MAP', 1234, 'testvrfname')
        self.cache.hset('VRFNAME_VRFID_MAP', 'testvrfname2', 1235)
        self.cache.hset('VRFID_VRFNAME_MAP', 1235, 'testvrfname2')

        r = self.put_config_interface_port_port('Ethernet12', {
            'vrf_id': 1234,
            'addr': '127.0.0.0/24',
            'mac_address': '01:02:03:04:05:06'
        })
        self.assertEqual(r.status_code, 204)

        r = self.put_config_interface_port_port('Ethernet12', {
            'vrf_id': 1235,
            'addr': '127.0.0.0/24',
            'mac_address': '02:02:03:04:05:06'
        })
        self.assertEqual(r.status_code, 405)

        keys = self.db.keys()
        self.assertEqual(sorted(keys), sorted([
            b'IGNORE_INTF_TABLE_KEY_SET',
            b'IGNORE_INTF_TABLE:Ethernet12:127.0.0.0/24',
            b'VLAN_MEMBER_TABLE:Vlan3840:Ethernet0',
            b'VLAN_MEMBER_TABLE_KEY_SET'
        ]))

        intf_table = self.db.hgetall('IGNORE_INTF_TABLE:Ethernet12:127.0.0.0/24')
        self.assertEqual(intf_table, {
            b'vrf_id': b'1234'
        })

        vlan_table = self.db.hgetall('VLAN_MEMBER_TABLE:Vlan3840:Ethernet0')
        self.assertEqual(vlan_table, {
            b'tagging_mode': b'tagged'
        })

        port_addr = self.cache.hget('PORT_ADDR_MAP', 'Ethernet12')
        self.assertEqual(b'127.0.0.0/24', port_addr)

        mac_addr = self.cache.hget('PORT_MAC_MAP', 'Ethernet12')
        self.assertEqual(b'01:02:03:04:05:06', mac_addr)

    def test_put_config_interface_port_port_spoof_guard_update(self):
        self.cache.hset('VRFNAME_VRFID_MAP', 'testvrfname', 1234)
        self.cache.hset('VRFID_VRFNAME_MAP', 1234, 'testvrfname')

        r = self.put_config_interface_port_port('Ethernet12', {
            'vrf_id': 1234,
            'addr': '127.0.0.1/24',
            'spoof_guard': ['127.0.0.1/24', '1.2.3.4/12', '5.4.3.2/18'],
            'mac_address': '01:02:03:04:05:06'
        })
        self.assertEqual(r.status_code, 204)

        r = self.put_config_interface_port_port('Ethernet12', {
            'vrf_id': 1234,
            'addr': '127.0.0.1/24',
            'spoof_guard': ['127.0.0.2/24', '1.2.3.5/12', '5.4.3.1/18'],
            'mac_address': '01:02:03:04:05:06'
        })
        self.assertEqual(r.status_code, 204)

        keys = self.db.keys()
        self.assertEqual(sorted(keys), sorted([
            b'IGNORE_INTF_TABLE_KEY_SET',
            b'IGNORE_INTF_TABLE:Ethernet12:127.0.0.1/24',
            b'SPOOF_GUARD_TABLE_KEY_SET',
            b'SPOOF_GUARD_TABLE:Ethernet12',
            b'VLAN_MEMBER_TABLE:Vlan3840:Ethernet0',
            b'VLAN_MEMBER_TABLE_KEY_SET'
        ]))

        intf_table = self.db.hgetall('IGNORE_INTF_TABLE:Ethernet12:127.0.0.1/24')
        self.assertEqual(intf_table, {
            b'vrf_id': b'1234'
        })

        vlan_table = self.db.hgetall('VLAN_MEMBER_TABLE:Vlan3840:Ethernet0')
        self.assertEqual(vlan_table, {
            b'tagging_mode': b'tagged'
        })

        spoof_guard_table = self.db.hgetall('SPOOF_GUARD_TABLE:Ethernet12')
        self.assertEqual(spoof_guard_table, {
            b'addr_list': b'127.0.0.2/24,1.2.3.5/12,5.4.3.1/18'
        })

        port_addr = self.cache.hget('PORT_ADDR_MAP', 'Ethernet12')
        self.assertEqual(b'127.0.0.1/24', port_addr)

        mac_addr = self.cache.hget('PORT_MAC_MAP', 'Ethernet12')
        self.assertEqual(b'01:02:03:04:05:06', mac_addr)

    def test_delete_config_interface_port_port(self):
        self.cache.hset('PORT_ADDR_MAP', 'Ethernet12', '127.0.0.0/24')
        self.cache.hset('PORT_MAC_MAP', 'Ethernet12', '01:02:03:04:05:06')
        self.db.hset('IGNORE_INTF_TABLE:Ethernet12:127.0.0.0/24', 'vrf_id', '1234')

        r = self.delete_config_interface_port_port('Ethernet12')
        self.assertEqual(r.status_code, 204)

    def test_get_config_interface_port_port(self):
        self.cache.hset('PORT_ADDR_MAP', 'Ethernet0', '127.0.0.0/24')
        self.cache.hset('PORT_MAC_MAP', 'Ethernet0', '01:02:03:04:05:06')
        self.db.hset('IGNORE_INTF_TABLE:Ethernet0:127.0.0.0/24', 'vrf_id', '1234')

        r = self.get_config_interface_port_port('Ethernet0')
        self.assertEqual(r.status_code, 200)

        j = json.loads(r.text)
        self.assertEqual(j, {
            'port': 'Ethernet0',
            'attr': {
                'addr': '127.0.0.0/24',
                'vrf_id': 1234,
                'mac_address':'01:02:03:04:05:06'
            }
        })

    # /state/interface/{port}
    def test_get_state_interface_port(self):
        r = self.get_state_interface_port('lo')
        self.assertEqual(r.status_code, 200)

    # /state/interface
    def test_get_state_interface(self):
        r = self.get_state_interface()
        self.assertEqual(r.status_code, 200)

    # /config/tunnel/decap/{tunnel_type}
    def test_put_config_tunnel_decap_tunnel_type(self):
        r = self.put_config_tunnel_decap_tunnel_type('vxlan', {
            'ip_addr': '34.53.1.0'
        })
        self.assertEqual(r.status_code, 204)

        keys = self.db.keys()
        self.assertEqual(sorted(keys), sorted([
            b'TUNNEL_TABLE:decapsulation:vxlan',
            b'TUNNEL_TABLE_KEY_SET',
            b'ACL_RULE_TABLE:DPDK:RULE_1',
            b'ACL_RULE_TABLE:DPDK:RULE_2',
            b'ACL_RULE_TABLE_KEY_SET',
            b'ACL_TABLE:DPDK',
            b'ACL_TABLE_KEY_SET'
        ]))

        tunnel_table = self.db.hgetall('TUNNEL_TABLE:decapsulation:vxlan')
        self.assertEqual(tunnel_table, {b'local_termination_ip': b'34.53.1.0'})

        acl_table = self.db.hgetall('ACL_TABLE:DPDK')
        self.assertEqual(acl_table, {
            b'policy_desc': b'dpdk',
            b'ports': b'Ethernet4,Ethernet8',
            b'type': b'L3'
        })

        acl_rule_table = self.db.hgetall('ACL_RULE_TABLE:DPDK:RULE_1')
        self.assertEqual(acl_rule_table, {
            b'dst_ip': b'127.0.0.1/32',
            b'l4_dst_port': b'4789',
            b'packet_action': b'REDIRECT:1.1.1.2',
            b'priority': b'9999'
        })

        acl_rule_table = self.db.hgetall('ACL_RULE_TABLE:DPDK:RULE_2')
        self.assertEqual(acl_rule_table, {
            b'dst_ip': b'127.0.0.1/32',
            b'l4_dst_port': b'65330',
            b'packet_action': b'REDIRECT:1.1.1.2',
            b'priority': b'9999'
        })

    def test_get_config_tunnel_decap_tunnel_type(self):
        self.db.hset('TUNNEL_TABLE:decapsulation:vxlan', b'local_termination_ip', b'34.53.1.0')

        r = self.get_config_tunnel_decap_tunnel_type('vxlan')
        self.assertEqual(r.status_code, 200)

        j = json.loads(r.text)
        self.assertEqual(j, {
            'tunnel_type': 'vxlan',
            'attr': {
                'ip_addr': '34.53.1.0'
            }
        })

    def test_delete_config_tunnel_decap_tunnel_type(self):
        self.db.hset('TUNNEL_TABLE:decapsulation:vxlan', b'local_termination_ip', b'34.53.1.0')
        self.db.sadd('TUNNEL_TABLE_KEY_SET', b'decapsulation:vxlan')
        self.db.hmset('ACL_TABLE:DPDK', {
            b'policy_desc': b'dpdk',
            b'ports': b'Ethernet0',
            b'type': b'L3'
        })
        self.db.hmset('ACL_RULE_TABLE:DPDK:RULE_1', {
            b'dst_ip': b'127.0.0.1/32',
            b'l4_dst_port': b'4789',
            b'packet_aciton': 'REDIRECT:Ethernet0',
            b'priority': b'9999'
        })
        self.db.hmset('ACL_RULE_TABLE:DPDK:RULE_2', {
            b'dst_ip': b'127.0.0.1/32',
            b'l4_dst_port': b'65330',
            b'packet_aciton': 'REDIRECT:Ethernet0',
            b'priority': b'9999'
        })
        self.db.sadd('ACL_TABLE_KEY_SET', b'DPDK')
        self.db.sadd('ACL_RULE_TABLE_KEY_SET', b'DPDK:RULE_1', b'DPDK:RULE_2')

        r = self.delete_config_tunnel_decap_tunnel_type('vxlan')
        self.assertEqual(r.status_code, 204)

        keys = self.db.keys()
        self.assertEqual(sorted(keys), sorted([
            b'TUNNEL_TABLE_KEY_SET',
            b'ACL_TABLE_KEY_SET',
            b'ACL_RULE_TABLE_KEY_SET'
        ]))

class msee_invalid_input_tests(msee_client):
    """Invalid input tests"""
    # /config/vrouter/{vrf_id}
    def test_put_config_vrouter_vrf_id_duplicate_name(self):
        self.cache.hset('VRFNAME_VRFID_MAP', 'testvrfname', 9876)
        self.cache.hset('VRFID_VRFNAME_MAP', 9876, 'testvrfname')

        r = self.put_config_vrouter_vrf_id(1234, {
            'vrf_name': 'testvrfname'
        })
        self.assertEqual(r.status_code, 400)

        j = json.loads(r.text)
        self.assertListEqual(['vrf_name'], j['error']['fields'])

    def test_put_config_vrouter_vrf_id_vrf_name_not_string(self):
        r = self.put_config_vrouter_vrf_id(1234, {
            'vrf_name': 0
        })
        self.assertEqual(r.status_code, 400)

        j = json.loads(r.text)
        self.assertListEqual(['vrf_name'], j['error']['fields'])

    def test_put_config_vrouter_vrf_id_vrf_name_missing(self):
        r = self.put_config_vrouter_vrf_id(1234, {})
        self.assertEqual(r.status_code, 400)

        j = json.loads(r.text)
        self.assertListEqual(['vrf_name'], j['error']['fields'])

    def test_get_config_vrouter_vrf_id_vrf_id_missing(self):
        r = self.get_config_vrouter_vrf_id(1234)
        self.assertEqual(r.status_code, 404)

    # /config/vrouter

    # /config/tunnel/encap/vxlan/{vnid}
    def test_put_config_tunnel_encap_vxlan_vnid_vrf_id_not_created(self):
        r = self.put_config_tunnel_encap_vxlan_vnid(1234, {
            'vrf_id': 9876
        })
        self.assertEqual(r.status_code, 404)

        j = json.loads(r.text)
        self.assertListEqual(['vrf_id'], j['error']['fields'])

    def test_put_config_tunnel_encap_vxlan_vnid_vrf_id_missing(self):
        r = self.put_config_tunnel_encap_vxlan_vnid(1234, {})
        self.assertEqual(r.status_code, 400)

        j = json.loads(r.text)
        self.assertListEqual(['vrf_id'], j['error']['fields'])

    def test_get_config_tunnel_encap_vxlan_vnid_missing_tunnel(self):
        r = self.get_config_tunnel_encap_vxlan_vnid(1234)
        self.assertEqual(r.status_code, 404)

    # /config/tunnel/encap/vxlan

    # /config/vrouter/{vrf_id}/routes
    def test_put_config_vrouter_vrf_id_routes_vnid_not_exist(self):
        self.cache.hset('VRFNAME_VRFID_MAP', 'testvrfname', 1234)
        self.cache.hset('VRFID_VRFNAME_MAP', 1234, 'testvrfname')

        r = self.put_config_vrouter_vrf_id_routes(1234, [
            {
                'ip_prefix': '127.0.0.0/24',
                'nexthop_type' : 'vxlan-tunnel',
                'nexthop': '92.89.1.2',
                'mac_address': '01:02:03:04:05:06',
                'vnid': 9876,
                'src_ip': '192.168.0.1',
                'error': 'none'
            }
        ])
        self.assertEqual(r.status_code, 201)

        j = json.loads(r.text)
        self.assertEqual(j, {
            'failed': [{
                'ip_prefix': '127.0.0.0/24',
                'nexthop_type' : 'vxlan-tunnel',
                'nexthop': '92.89.1.2',
                'mac_address': '01:02:03:04:05:06',
                'vnid': 9876,
                'src_ip': '192.168.0.1',
                'port': 'azure',
                'error': 'none'
            }]
        })

    def test_put_config_vrouter_vrf_id_routes_vrfid_not_created(self):
        self.cache.hset('VRFID_VLANID_MAP', 1234, 2)
        self.cache.setbit('VLANID_ALLOC', 0, 1)

        r = self.put_config_vrouter_vrf_id_routes(1234, [
            {
                'ip_prefix': '127.0.0.0/24',
                'nexthop_type' : 'vxlan-tunnel',
                'nexthop': '92.89.1.2',
                'mac_address': '01:02:03:04:05:06',
                'vnid': 9876,
                'src_ip': '192.168.0.1',
                'error': 'none'
            }
        ])
        self.assertEqual(r.status_code, 404)

        j = json.loads(r.text)
        self.assertListEqual(['vrf_id'], j['error']['fields'])

    def test_put_config_vrouter_vrf_id_routes_mac_missing(self):
        r = self.put_config_vrouter_vrf_id_routes(1234, [
            {
                'ip_prefix': '127.0.0.0/24',
                'nexthop_type' : 'vxlan-tunnel',
                'nexthop': '92.89.1.2',
                'vnid': 9876,
                'src_ip': '192.168.0.1',
                'error': 'none'
            }
        ])
        self.assertEqual(r.status_code, 400)

        j = json.loads(r.text)
        self.assertListEqual(['mac_address'], j['error']['fields'])

    def test_put_config_vrouter_vrf_id_routes_ip_prefix_missing(self):
        r = self.put_config_vrouter_vrf_id_routes(1234, [
            {
                'nexthop_type' : 'vxlan-tunnel',
                'nexthop': '92.89.1.2',
                'mac_address': '01:02:03:04:05:06',
                'vnid': 9876,
                'src_ip': '192.168.0.1',
                'error': 'none'
            }
        ])
        self.assertEqual(r.status_code, 400)

        j = json.loads(r.text)
        self.assertListEqual(['ip_prefix'], j['error']['fields'])

    def test_put_config_vrouter_vrf_id_routes_nexthop_type_missing(self):
        r = self.put_config_vrouter_vrf_id_routes(1234, [
            {
                'ip_prefix': '127.0.0.0/24',
                'nexthop': '92.89.1.2',
                'vnid': 9876,
                'src_ip': '192.168.0.1',
                'error': 'none'
            }
        ])
        self.assertEqual(r.status_code, 400)

        j = json.loads(r.text)
        self.assertListEqual(['nexthop_type'], j['error']['fields'])

    def test_put_config_vrouter_vrf_id_routes_nexthop_missing(self):
        r = self.put_config_vrouter_vrf_id_routes(1234, [
            {
                'ip_prefix': '127.0.0.0/24',
                'nexthop_type' : 'vxlan-tunnel',
                'mac_address': '01:02:03:04:05:06',
                'vnid': 9876,
                'src_ip': '192.168.0.1',
                'error': 'none'
            }
        ])
        self.assertEqual(r.status_code, 400)

        j = json.loads(r.text)
        self.assertListEqual(['nexthop'], j['error']['fields'])

    def test_put_config_vrouter_vrf_id_routes_nexthop_type_invalid(self):
        r = self.put_config_vrouter_vrf_id_routes(1234, [
            {
                'ip_prefix': '127.0.0.0/24',
                'nexthop_type' : 'notvalid',
                'nexthop': '92.89.1.2',
                'vnid': 9876,
                'src_ip': '192.168.0.1',
                'error': 'none'
            }
        ])
        self.assertEqual(r.status_code, 400)

        j = json.loads(r.text)
        self.assertListEqual(['nexthop_type'], j['error']['fields'])

    # /config/interface/qinq/{port}/{stag}/{ctag}
    def test_put_config_interface_qinq_port_stag_ctag_vrf_id_missing(self):
        r = self.put_config_interface_qinq_port_stag_ctag(42, 5, 6, {
            'peer_ip': '1.2.3.4',
            'proxy_arp_ip': '5.6.7.8',
            'subnet': '1.2.3.0/24'
        })
        self.assertEqual(r.status_code, 400)

        j = json.loads(r.text)
        self.assertListEqual(['vrf_id'], j['error']['fields'])

    def test_put_config_interface_qinq_port_stag_ctag_ctag_outofrange(self):
        r = self.put_config_interface_qinq_port_stag_ctag(42, 5, 6000, {
            'vrf_id': 1234,
            'peer_ip': '1.2.3.4',
            'proxy_arp_ip': '5.6.7.8',
            'subnet': '1.2.3.0/24'
        })
        self.assertEqual(r.status_code, 400)

        j = json.loads(r.text)
        self.assertListEqual(['ctag'], j['error']['fields'])

    def test_put_config_interface_qinq_port_stag_ctag_stag_outofrange(self):
        r = self.put_config_interface_qinq_port_stag_ctag(42, 1, 3000, {
            'vrf_id': 1234,
            'peer_ip': '1.2.3.4',
            'proxy_arp_ip': '5.6.7.8',
            'subnet': '1.2.3.0/24'
        })
        self.assertEqual(r.status_code, 400)

        j = json.loads(r.text)
        self.assertListEqual(['stag'], j['error']['fields'])

    def test_put_config_interface_qinq_port_stag_ctag_ipv6(self):
        r = self.put_config_interface_qinq_port_stag_ctag(42, 5, 6, {
            'vrf_id': 1234,
            'peer_ip': '2620:0:ccc::2',
            'proxy_arp_ip': '5.6.7.8',
            'subnet': '1.2.3.0/24'
        })
        self.assertEqual(r.status_code, 400)

        j = json.loads(r.text)
        self.assertListEqual(['peer_ip'], j['error']['fields'])

    def test_put_config_interface_qinq_port_stag_ctag_cidr_ipv6(self):
        r = self.put_config_interface_qinq_port_stag_ctag(42, 5, 6, {
            'vrf_id': 1234,
            'peer_ip': '1.2.3.4',
            'proxy_arp_ip': '5.6.7.8',
            'subnet': '2620:0:ccc::2/32'
        })
        self.assertEqual(r.status_code, 400)

        j = json.loads(r.text)
        self.assertListEqual(['subnet'], j['error']['fields'])

    def test_get_config_interface_qinq_port_stag_ctag_missing(self):
        r = self.get_config_interface_qinq_port_stag_ctag('Ethernet0', 5, 6)
        self.assertEqual(r.status_code, 404)

    # /config/interface/qinq/{port}

    # /config/interface/port/{port}
    def test_put_config_interface_port_port_vrf_id_not_created(self):
        r = self.put_config_interface_port_port('Ethernet0', {
            'vrf_id': 1234,
            'addr': '127.0.0.1/24',
            'mac_address': '01:02:03:04:05:06'
        })
        self.assertEqual(r.status_code, 404)

        j = json.loads(r.text)
        self.assertListEqual(['vrf_id'], j['error']['fields'])

    def test_put_config_interface_port_port_vrf_id_missing(self):
        r = self.put_config_interface_port_port('Ethernet0', {
            'addr': '127.0.0.1/24',
            'mac_address': '01:02:03:04:05:06'
        })
        self.assertEqual(r.status_code, 400)

        j = json.loads(r.text)
        self.assertListEqual(['vrf_id'], j['error']['fields'])

    def test_put_config_interface_port_port_mac_missing(self):
        r = self.put_config_interface_port_port('Ethernet0', {
            'vrf_id': 1234,
            'addr': '127.0.0.1/24'
        })
        self.assertEqual(r.status_code, 400)

        j = json.loads(r.text)
        self.assertListEqual(['mac_address'], j['error']['fields'])

    def test_put_config_interface_port_port_ip_ipv6(self):
        r = self.put_config_interface_port_port('Ethernet0', {
            'vrf_id': 1234,
            'addr': '2620:0:ccc::2/24',
            'mac_address': '01:02:03:04:05:06'
        })
        self.assertEqual(r.status_code, 400)

        j = json.loads(r.text)
        self.assertListEqual(['addr'], j['error']['fields'])

    def test_delete_config_interface_port_port_port_not_created(self):
        r = self.delete_config_interface_port_port('Ethernet0')
        self.assertEqual(r.status_code, 404)

        j = json.loads(r.text)
        self.assertListEqual(['port'], j['error']['fields'])

    def test_get_config_interface_port_port_port_not_created(self):
        r = self.get_config_interface_port_port('Ethernet0')
        self.assertEqual(r.status_code, 404)

        j = json.loads(r.text)
        self.assertListEqual(['port'], j['error']['fields'])

    # /state/interface/{port}
    def test_get_state_interface_port_port_missing(self):
        r = self.get_state_interface_port('this_is_not_a_real_port')
        self.assertEqual(r.status_code, 404)

        j = json.loads(r.text)
        self.assertListEqual(['port'], j['error']['fields'])

    # /state/interface

    # /config/tunnel/decap/{tunnel_type}
    def test_put_config_tunnel_decap_tunnel_type_ip_out_of_range(self):
        r = self.put_config_tunnel_decap_tunnel_type('vxlan', {
            'ip_addr': '256.53.1.0'
        })
        self.assertEqual(r.status_code, 400)

        j = json.loads(r.text)
        self.assertListEqual(['ip_addr'], j['error']['fields'])

    def test_put_config_tunnel_decap_tunnel_type_ip_missing_segment(self):
        r = self.put_config_tunnel_decap_tunnel_type('vxlan', {
            'ip_addr': '255.53.1'
        })
        self.assertEqual(r.status_code, 400)

        j = json.loads(r.text)
        self.assertListEqual(['ip_addr'], j['error']['fields'])

    def test_put_config_tunnel_decap_tunnel_type_ipv6(self):
        r = self.put_config_tunnel_decap_tunnel_type('vxlan', {
            'ip_addr': '2620:0:ccc::2'
        })
        self.assertEqual(r.status_code, 400)

        j = json.loads(r.text)
        self.assertListEqual(['ip_addr'], j['error']['fields'])

    def test_put_config_tunnel_decap_tunnel_type_ip_not_string(self):
        r = self.put_config_tunnel_decap_tunnel_type('vxlan', {
            'ip_addr': 0
        })
        self.assertEqual(r.status_code, 400)

        j = json.loads(r.text)
        self.assertListEqual(['ip_addr'], j['error']['fields'])

    def test_put_config_tunnel_decap_tunnel_type_ip_missing(self):
        r = self.put_config_tunnel_decap_tunnel_type('vxlan', {})
        self.assertEqual(r.status_code, 400)

        j = json.loads(r.text)
        self.assertListEqual(['ip_addr'], j['error']['fields'])

    def test_put_config_tunnel_decap_tunnel_type_not_vxlan(self):
        r = self.put_config_tunnel_decap_tunnel_type('not_vxlan', {
            'ip_addr': '192.128.0.1'
        })
        self.assertEqual(r.status_code, 400)

        j = json.loads(r.text)
        self.assertListEqual(['tunnel_type'], j['error']['fields'])

    def test_delete_config_tunnel_decap_tunnel_type_not_vxlan(self):
        r = self.delete_config_tunnel_decap_tunnel_type('not_vxlan')
        self.assertEqual(r.status_code, 400)

        j = json.loads(r.text)
        self.assertListEqual(['tunnel_type'], j['error']['fields'])

    def test_get_config_tunnel_decap_tunnel_type_no_tunnel(self):
        r = self.get_config_tunnel_decap_tunnel_type('vxlan')
        self.assertEqual(r.status_code, 404)

if __name__ == '__main__':
    unittest.main()
