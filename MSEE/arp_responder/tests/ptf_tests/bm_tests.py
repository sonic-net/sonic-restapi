import ptf
from ptf.base_tests import BaseTest
import ptf.testutils as testutils
from ptf.testutils import simple_arp_packet
from pprint import pprint


class BMEasyTest(BaseTest):
    def setUp(self):
        self.dataplane = ptf.dataplane_instance
        self.dataplane.flush()
        self.my_mac = {}
        self.remote_mac = {}
        for port_id, port in self.dataplane.ports.iteritems():
            self.my_mac[port_id[1]] = port.mac()
            self.remote_mac[port_id[1]] = self.get_remote_mac(port_id[1])

    def tearDown(self):
        pass

    def runTest(self):
        self.test_port(0)
        self.test_port(1)

    def test_port(self, port_number):
        src_mac = self.my_mac[port_number]
        packet = simple_arp_packet(
                          eth_src=src_mac,
                          vlan_vid=0,
                          vlan_pcp=0,
                          ip_snd='192.168.0.1',
                          ip_tgt='192.168.0.2',
                          hw_snd=src_mac,
                          hw_tgt='00:00:00:00:00:00',
                          )
        testutils.send_packet(self, (0, port_number), packet)

        exp_packet = simple_arp_packet(
                          pktlen=42,
                          eth_dst=src_mac,
                          eth_src=self.remote_mac[port_number],
                          vlan_vid=0,
                          vlan_pcp=0,
                          arp_op=2,
                          ip_snd='192.168.0.2',
                          ip_tgt='192.168.0.1',
                          hw_snd=self.remote_mac[port_number],
                          hw_tgt=src_mac
                          )
        testutils.verify_packet(self, exp_packet, port_number)

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
        
