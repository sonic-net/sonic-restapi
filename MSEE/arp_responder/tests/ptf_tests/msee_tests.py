import sys
import ptf
from ptf.base_tests import BaseTest
import ptf.testutils as testutils
from ptf.testutils import simple_arp_packet
import struct
from threading import Thread
from pprint import pprint

sys.path.append('/usr/lib/python2.7/site-packages')

from thrift import Thrift
from thrift.transport import TSocket
from thrift.transport import TTransport
from thrift.protocol import TBinaryProtocol

from arp_responder import arp_responder
from arp_responder.ttypes import req_tuple_t, req_tuples_t, rep_tuple_t


class MSEEEasyTest(BaseTest):
    def setUp(self):
        self.dataplane = ptf.dataplane_instance
        self.dataplane.flush()
        self.my_mac = {}
        self.remote_mac = {}
        for port_id, port in self.dataplane.ports.iteritems():
            self.my_mac[port_id[1]] = port.mac()
            self.remote_mac[port_id[1]] = self.get_remote_mac(port_id[1])

        self.result = None

        socket = TSocket.TSocket('localhost', 9091)
        self.transport = TTransport.TBufferedTransport(socket)
        protocol = TBinaryProtocol.TBinaryProtocol(self.transport)
        self.client = arp_responder.Client(protocol)
        self.transport.open()

    def tearDown(self):
        self.transport.close()

    def runTest(self):
        self.test_reply()
        self.test_request()

    def test_reply(self):
        self.test_reply_qinq(0)
        self.test_reply_qinq(1)

    def test_reply_qinq(self, port_number):
        intf = 'iif%d' % port_number
        stag = 10
        ctag = 20
        self.client.add_interface(intf)
        self.client.add_ip(intf, stag, ctag, 0x01020304)
        src_mac = self.my_mac[port_number]
        packet = simple_arp_packet(
                          pktlen=42,
                          eth_src=src_mac,
                          ip_snd='1.2.3.1',
                          ip_tgt='1.2.3.4',
                          hw_snd=src_mac,
                          hw_tgt='00:00:00:00:00:00',
                          )
        tagged_packet = self.insert_tags(packet, stag, ctag)
        testutils.send_packet(self, (0, port_number), tagged_packet)
        self.client.del_ip(intf, stag, ctag)
        self.client.del_interface(intf)

        exp_packet = simple_arp_packet(
                          pktlen=42,
                          eth_dst=src_mac,
                          eth_src=self.remote_mac[port_number],
                          arp_op=2,
                          ip_snd='1.2.3.4',
                          ip_tgt='1.2.3.1',
                          hw_snd=self.remote_mac[port_number],
                          hw_tgt=src_mac
                          )
        tagged_exp_packet = self.insert_tags(exp_packet, stag, ctag)
        testutils.verify_packet(self, tagged_exp_packet, port_number)

    def test_request(self):
        thr = Thread(target=self.request_mac_thread)
        thr.start()
        stag = 10
        ctag = 20
        exp_packet_1 = simple_arp_packet(
                          pktlen=42,
                          eth_dst='ff:ff:ff:ff:ff:ff',
                          eth_src=self.remote_mac[0],
                          ip_snd='1.2.3.4',
                          ip_tgt='1.2.3.1',
                          hw_snd=self.remote_mac[0],
                          hw_tgt='ff:ff:ff:ff:ff:ff'
                          )
        t_exp_packet0 = self.insert_tags(exp_packet_1, stag, ctag)
        testutils.verify_packet(self, t_exp_packet0, 0)
        exp_packet_2 = simple_arp_packet(
                          pktlen=42,
                          eth_dst='ff:ff:ff:ff:ff:ff',
                          eth_src=self.remote_mac[1],
                          ip_snd='1.2.3.5',
                          ip_tgt='1.2.3.1',
                          hw_snd=self.remote_mac[1],
                          hw_tgt='ff:ff:ff:ff:ff:ff'
                          )
        t_exp_packet1 = self.insert_tags(exp_packet_2, stag, ctag)
        testutils.verify_packet(self, t_exp_packet1, 1)

        packet = simple_arp_packet(
                          pktlen=42,
                          eth_dst=self.remote_mac[0],
                          eth_src=self.my_mac[0],
                          arp_op=2,
                          ip_snd='1.2.3.1',
                          ip_tgt='1.2.3.4',
                          hw_snd=self.my_mac[0],
                          hw_tgt=self.remote_mac[0]
                          )
        tagged_packet = self.insert_tags(packet, stag, ctag)
        testutils.send_packet(self, (0, 0), tagged_packet)
        thr.join()
        result_mac = ":".join("%02x" % v for v in list(struct.unpack("BBBBBB", self.result[0].mac)))
        self.assertTrue(self.result[0].index == 0)
        self.assertTrue(self.result[0].is_found)
        self.assertTrue(result_mac == self.my_mac[0])
        self.assertTrue(self.result[0].request.stag == 10)
        self.assertTrue(self.result[0].request.ctag == 20)
        self.assertTrue(self.result[0].request.iface_name == 'iif0')

    def request_mac_thread(self):
        self.client.add_interface('iif0')
        self.client.add_ip('iif0', 10, 20, 0x01020304)
        self.client.add_interface('iif1')
        self.client.add_ip('iif1', 10, 20, 0x01020305)

        t1 = req_tuples_t([req_tuple_t('iif0', 10, 20), req_tuple_t('iif1', 10, 20)], 0, 0x01020301)
        self.result = self.client.request_mac([t1])

        self.client.del_ip('iif1', 10, 20)
        self.client.del_interface('iif1')
        self.client.del_ip('iif0', 10, 20)
        self.client.del_interface('iif0')

    def get_remote_mac(self, port_number):
        mac, _, _ = self.cmd(['docker', 'exec', '-ti', 'arpresponder_test', 'cat', '/sys/class/net/iif%d/address' % port_number])
        return mac.strip()

    def cmd(self, cmds):
        process = subprocess.Popen(cmds,
                                   shell=False,
                                   stdout=subprocess.PIPE,
                                   stderr=subprocess.PIPE)
        stdout, stderr = process.communicate()
        return_code = process.returncode

        return stdout, stderr, return_code

    def insert_tags(self, packet, stag, ctag):
        p = str(packet)
        vlan_hdr = struct.pack("!HHHH",0x88A8, stag, 0x8100, ctag)
        return p[0:12] + vlan_hdr + p[12:]


