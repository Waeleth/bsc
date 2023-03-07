// Copyright 2015 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.

// bootnode runs a bootstrap node for the Ethereum Discovery Protocol.
package main

import (
	"crypto/ecdsa"
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/nat"
	"github.com/ethereum/go-ethereum/p2p/netutil"
)

func main() {
	var (
		listenAddr    = flag.String("addr", ":30301", "listen address")
		genKey        = flag.String("genkey", "", "generate a node key")
		writeAddr     = flag.Bool("writeaddress", false, "write out the node's public key and quit")
		nodeKeyFile   = flag.String("nodekey", "", "private key filename")
		nodeKeyHex    = flag.String("nodekeyhex", "", "private key as hex (for testing)")
		natdesc       = flag.String("nat", "none", "port mapping mechanism (any|none|upnp|pmp|extip:<IP>)")
		netrestrict   = flag.String("netrestrict", "", "restrict network communication to the given IP networks (CIDR masks)")
		runv5         = flag.Bool("v5", false, "run a v5 topic discovery bootnode")
		verbosity     = flag.Int("verbosity", int(log.LvlInfo), "log verbosity (0-5)")
		vmodule       = flag.String("vmodule", "", "log verbosity pattern")
		networkFilter = flag.String("network", "", "<bsc/chapel/rialto/yolo> filters nodes by eth ENR entry")

		nodeKey        *ecdsa.PrivateKey
		filterFunction discover.NodeFilter
		err            error
	)

	var BSCBootnodes = []string{
		// Ethereum Foundation Go Bootnodes
		"enode://1cc4534b14cfe351ab740a1418ab944a234ca2f702915eadb7e558a02010cb7c5a8c295a3b56bcefa7701c07752acd5539cb13df2aab8ae2d98934d712611443@52.71.43.172:30311",
		"enode://28b1d16562dac280dacaaf45d54516b85bc6c994252a9825c5cc4e080d3e53446d05f63ba495ea7d44d6c316b54cd92b245c5c328c37da24605c4a93a0d099c4@34.246.65.14:30311",
		"enode://5a7b996048d1b0a07683a949662c87c09b55247ce774aeee10bb886892e586e3c604564393292e38ef43c023ee9981e1f8b335766ec4f0f256e57f8640b079d5@35.73.137.11:30311",
		"enode://9f90d69c5fef1ca0b1417a1423038aa493a7f12d8e3d27e10a5a8fd3da216e485cf6c15f48ee310a14729bc3a4b05038479476c0aa82eed3c5d9d2e64ba3a2b3@52.69.42.169:30311", "enode://78ef719ebb2f4fc222aa988a356274dcd3624fb808936ca2ea77388ca229773d4351f795abf505e86db1a30ed1523ded9f9674d916b295bfb98516b78d2844be@13.231.200.147:30311", "enode://a8ff9670029785a644fb709ec7cd7e7e2d2b93761872bfe1b011a1ed1c601b23ffa69ead0901b759d780ed65aa81444261905b6964bdf8647bf5b061a4796d2d@54.168.191.244:30311", "enode://0f0abad52d6e3099776f70fda913611ad33c9f4b7cafad6595691ea1dd57a37804738be65315fc417d41ab52632c55a5f5f1e5ed3123ed64a312341a8c3f9e3c@52.193.230.222:30311", "enode://ecc277f466f35b249b62de8ca567dfe759162ffecc79f40339655537ee58132aec892bc0c4ad3dfb0ba5441bb7a68301c0c09e3f66454110c2c03ccca084c6b5@54.238.240.9:30311", "enode://dd3fb5f4da631067d0a9206bb0ac4400d3a076102194257911b632c5aa56f6a3289a855cc0960ad7f2cda3ba5162e0d879448775b07fa73ccd2e4e0477290d9a@54.199.96.72:30311", "enode://74481dd5079320755588b5243f82ddec7364ad36108ac77272b8e003194bb3f5e6386fcd5e50a0604db1032ac8cb9b58bb813f8e57125ad84ec6ceec65d29b4b@52.192.99.35:30311", "enode://190df80c16509d9d205145495f169a605d1459e270558f9684fcd7376934e43c65a38999d5e49d2ad118f49abfb6ff62068051ce49acc029da7d2be9910fe9fd@13.113.113.139:30311", "enode://368fc439d8f86f459822f67d9e8d1984bab32098096dc13d4d361f8a4eaf8362caae3af86e6b31524bda9e46910ac61b075728b14af163eca45413421386b7e2@52.68.165.102:30311", "enode://2038dac8d835db7c4c1f9d2647e37e6f5c5dc5474853899adb9b61700e575d237156539a720ff53cdb182ee56ac381698f357c7811f8eadc56858e0d141dcce0@18.182.11.67:30311", "enode://fc0bb7f6fc79ad7d867332073218753cb9fe5687764925f8405459a98b30f8e39d4da3a10f87fe06aa10df426c2c24c3907a4d81df4e3c88e890f7de8f8980de@54.65.239.152:30311", "enode://3aaaa0e0c7961ef3a9bf05f879f84308ca59651327cf94b64252f67448e582dcd6a6dbe996264367c8aa27fc302736db0283a3516c7406d48f268c5e317b9d49@34.250.1.192:30311", "enode://62c516645635f0389b4c851bfc4545720fac0607de74942e4ea7e923f4fa2ac0c438c146e2f0721c8ce06dca4e7f30f5c0136569d9f4b6a827c62b980fd53272@52.215.57.20:30311", "enode://5df2f71ae6b2e3bb92f92badbce1f601feabd2d6ce899cf8265c39c38ff446136d74f5bfa089532c7074bb7606a509a54a2ac66397aaaab2363dad3f43c687a8@79.125.103.83:30311", "enode://760b5fde9bc14155fa2a87e56cf610701ad6c1adcf44555a7b839baf71f86f11cdadcaf925e50b17c98cc28e20e0df3c3463caad7c6658a76ab68389af639f33@34.243.1.225:30311", "enode://57824d2d9b5f39681bee265d56ec98a17fa4af343debdeba18596837f776f7c6370d8a33354e2b1750c41b221778e05c4189b93aca0d4cb1d45d32dc3b2d63f1@34.240.198.163:30311", "enode://9b7ff9e2d2154f6de3f53db2123e6f9a6b5b29414d9d5ae8277592b361158c25fcab86e6bfad5ef6554c6d92fb4ca897f7342563e355b80bcdc994f9c268dc2f@34.251.95.115:30311", "enode://67ec1f3df346e0aef401175119172e86a20e7ee1442cba4a2074519405cdae3708be3fdcb5e139094408b5d6f6c8e85f89ebb77d04833f7aa251c91344dbd4c9@3.249.178.199:30311", "enode://99c8d55d4528330fc494705ea15c2a8be9c25cb638ed561657a642d57e7851e38365d20b6864419e82e593e2b8d22cee23a09e9bb774ec8f15795b077bae7aeb@54.229.26.251:30311", "enode://1afc9727301dcd8d2c5aef067031639ae3d3c7a23f8ba6c588a6a1b2c3cbcd738b4ccc53c07d08690ef591b99fd12f00a005f38d820354a91f418ab0939b9072@34.253.216.225:30311", "enode://7c7b46ad65325f16768013167a0b2ca3eaa20e5d594011b6202b9c4707f740e2c795e84563b3a8c7986fdfb413ce88726a096f3cac8366ac9ebf073095c20584@34.243.12.13:30311", "enode://71ef36f019bbdaa2a7b4676a61d014d0be81958e2c60dd95c66a5e1af10de6f3a62ecf9ad0c26b6c5789b81ac22f774abb4735cd9e259185773ebfd1efded5de@54.170.254.50:30311", "enode://627a1cb2c4712cce439026da0c2f599b97628c90c8ccc55526574a944b7455827544130b3003e79399cd79bd73a06a1d6bbd018fcf9ffc5297d3b731aa1b40ab@3.91.73.29:30311", "enode://16c7e98f78017dafeaa4129647d1ec66b32ee9be5ec753708820b7363091ceb310f575e7abd9603005e0e34d7b3316c1a4b6c8c42d7f074ed2eb4d073f800a03@3.85.216.212:30311", "enode://accbc0a5af0af03e1ec3b5e80544bdceea48011a6928cd82d2c1a9c38b65fd48ec970ba17bd8c0b0ec21a28faec9efe1d1ce55134784b9207146e2f62d8932ba@54.162.32.1:30311", "enode://c64c864572dae7ea25225a412c026ced0de66ae429b40c545be8f524a1aeb70b3441710dbfed19e3ba9ef08ce13b00a58daa7a7510924da8e6f4f412d8b45fd5@3.92.160.2:30311", "enode://5a838185d4b91eb42cbe3a60bb9f706484d8ec5041fa97b557d10e8ca10a459db0271e06e8b85cad57f1d2c7b05aa4319c0300b2936eefcb2302e10b253cf7d6@23.20.67.34:30311", "enode://3438d60bcb628ba33b0adf5e653751436fdc393a869fab136dec5ec6b2ed06d8ea30e4fec061f4f4a67bb01644897dbc3d14db44afc052eb69f102340aff70f9@18.215.252.114:30311", "enode://c307b4cddec0aea2188eafddedb0a076b9289402c63217b4c81eb7f34761c7cfaf6b075e93d7357169e226ff1bb4aa3bd71869b4c76cf261e2991005ddb4d4aa@3.81.81.182:30311", "enode://80f446f15c3c17b2f8cd7e0f7811f9ba62381abeabc0ce562134d6ac7d400aef212020c439f462d760ca250e8f14b50f215d65e7137d2e3e25d22dc8ff21bda7@54.162.73.225:30311", "enode://d69853daf3057cc191514afdf56df4769238fde4f261fab80c6e089480abb9916d61180e783d1cc9e5ae56d30ce6261d9954702dc73c41cd47e4b3961830b2dc@184.73.34.17:30311", "enode://ba88d1a8a5e849bec0eb7df9eabf059f8edeae9a9eb1dcf51b7768276d78b10d4ceecf0cde2ef191ced02f66346d96a36ca9da7d73542757d9677af8da3bad3f@54.198.97.197:30311", "enode://a232f92d1e76447b93306ece2f6a55ac70ca4633fae0938d71a100757eaf8526e6bbf720aa70cba1e6d186be17291ad1ee851a35596ec6caa2fdf135ce4b6b68@107.20.124.16:30311", "enode://2d55e48679442a9e3ef2a3edf2854dcb289f8162d57dbda1e82e7576b0708e0670befaa7255f5c9fa8389443a7e7b4ff762c9e7fd33ddf9f21ec9562f03e8945@18.212.135.123:30311", "enode://f7dc512940ca4a8f6858632abbdfc59cea6c4ed7a8da41ddfc4e4dac74e2664e74355fd7c688b285a22295e0053a800f759c9123ec741285a5bd602f89720cea@54.198.51.232:30311", "enode://9df97e190f0b82ba7891e0ed556f11f4c1a172c26b2e823e52cfe5722b3df3f1819d2acb87ed0bfeb21fe3aee4ef1ffb8c9227fa7fdf744bfd4f47caad461edf@54.81.89.198:30311",
	}
	flag.Parse()

	glogger := log.NewGlogHandler(log.StreamHandler(os.Stderr, log.TerminalFormat(false)))
	glogger.Verbosity(log.Lvl(*verbosity))
	glogger.Vmodule(*vmodule)
	log.Root().SetHandler(glogger)

	natm, err := nat.Parse(*natdesc)
	if err != nil {
		utils.Fatalf("-nat: %v", err)
	}
	switch {
	case *genKey != "":
		nodeKey, err = crypto.GenerateKey()
		if err != nil {
			utils.Fatalf("could not generate key: %v", err)
		}
		if err = crypto.SaveECDSA(*genKey, nodeKey); err != nil {
			utils.Fatalf("%v", err)
		}
		if !*writeAddr {
			return
		}
	case *nodeKeyFile == "" && *nodeKeyHex == "":
		utils.Fatalf("Use -nodekey or -nodekeyhex to specify a private key")
	case *nodeKeyFile != "" && *nodeKeyHex != "":
		utils.Fatalf("Options -nodekey and -nodekeyhex are mutually exclusive")
	case *nodeKeyFile != "":
		if nodeKey, err = crypto.LoadECDSA(*nodeKeyFile); err != nil {
			utils.Fatalf("-nodekey: %v", err)
		}
	case *nodeKeyHex != "":
		if nodeKey, err = crypto.HexToECDSA(*nodeKeyHex); err != nil {
			utils.Fatalf("-nodekeyhex: %v", err)
		}
	}

	if *networkFilter != "" {
		if filterFunction, err = discover.ParseEthFilter(*networkFilter); err != nil {
			utils.Fatalf("-network: %v", err)
		}
	}

	if *writeAddr {
		fmt.Printf("%x\n", crypto.FromECDSAPub(&nodeKey.PublicKey)[1:])
		os.Exit(0)
	}

	var restrictList *netutil.Netlist
	if *netrestrict != "" {
		restrictList, err = netutil.ParseNetlist(*netrestrict)
		if err != nil {
			utils.Fatalf("-netrestrict: %v", err)
		}
	}

	addr, err := net.ResolveUDPAddr("udp", *listenAddr)
	if err != nil {
		utils.Fatalf("-ResolveUDPAddr: %v", err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		utils.Fatalf("-ListenUDP: %v", err)
	}

	realaddr := conn.LocalAddr().(*net.UDPAddr)
	if natm != nil {
		if !realaddr.IP.IsLoopback() {
			go nat.Map(natm, nil, "udp", realaddr.Port, realaddr.Port, "ethereum discovery")
		}
		if ext, err := natm.ExternalIP(); err == nil {
			realaddr = &net.UDPAddr{IP: ext, Port: realaddr.Port}
		}
	}

	printNotice(&nodeKey.PublicKey, *realaddr)

	nodes := make([]*enode.Node, len(BSCBootnodes))
	for i, url := range BSCBootnodes {
		var err error
		nodes[i], err = enode.Parse(enode.ValidSchemes, url)
		if err != nil {
			panic("invalid node URL: " + err.Error())
		}
	}

	db, _ := enode.OpenDB("")
	ln := enode.NewLocalNode(db, nodeKey)
	cfg := discover.Config{
		PrivateKey:     nodeKey,
		NetRestrict:    restrictList,
		FilterFunction: filterFunction,
		Bootnodes:      nodes,
	}
	if *runv5 {
		if _, err := discover.ListenV5(conn, ln, cfg); err != nil {
			utils.Fatalf("%v", err)
		}
	} else {
		if _, err := discover.ListenUDP(conn, ln, cfg); err != nil {
			utils.Fatalf("%v", err)
		}
	}

	select {}
}

func printNotice(nodeKey *ecdsa.PublicKey, addr net.UDPAddr) {
	if addr.IP.IsUnspecified() {
		addr.IP = net.IP{127, 0, 0, 1}
	}
	n := enode.NewV4(nodeKey, addr.IP, 0, addr.Port)
	fmt.Println(n.URLv4())
	fmt.Println("Note: you're using cmd/bootnode, a developer tool.")
	fmt.Println("We recommend using a regular node as bootstrap node for production deployments.")
}
