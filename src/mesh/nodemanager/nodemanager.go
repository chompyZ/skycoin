package nodemanager

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/skycoin/skycoin/src/cipher"
	"github.com/skycoin/skycoin/src/mesh/domain"
	mesh "github.com/skycoin/skycoin/src/mesh/node"
	"github.com/skycoin/skycoin/src/mesh/transport"
)

func init() {
	ServerConfig = CreateTestConfig(15101)

	b := []byte{234, 15, 123, 220, 185, 171, 218, 20, 130, 48, 24, 255, 214, 133, 191, 164, 211, 190, 224, 127, 105, 125, 141, 178, 226, 250, 123, 149, 229, 33, 187, 165, 27}
	p := cipher.PubKey{}
	copy(p[:], b[:])
	ServerConfig.Node.PubKey = p

	//b2 := [32]byte{48, 126, 177, 168, 139, 146, 205, 8, 191, 110, 195, 254, 184, 22, 168, 118, 237, 126, 87, 224, 171, 243, 239, 87, 106, 152, 251, 217, 120, 239, 88, 138}
	//ServerConfig.Node.ChaCha20Key = b2
}

var ServerConfig *mesh.TestConfig

type NodeManager struct {
	ConfigList map[cipher.PubKey]*mesh.TestConfig
	Port       int
	NodesList  map[cipher.PubKey]*mesh.Node
	PubKeyList []cipher.PubKey
	Routes     map[RouteKey]Route
}

//configuration in here
type NodeManagerConfig struct {
}

//add config eventually
func NewNodeManager(config *NodeManagerConfig) *NodeManager {
	nm := NodeManager{}
	return &nm
}

// run node manager, in goroutine;
// call Shutdown to stop
func (self *NodeManager) Start() {

}

//called to trigger shutdown
func (self *NodeManager) Shutdown() {

}

type RouteKey struct {
	SourceNode cipher.PubKey
	TargetNode cipher.PubKey
}

type Route struct {
	SourceNode        cipher.PubKey
	TargetNode        cipher.PubKey
	RoutesToEstablish []cipher.PubKey
	Weight            int
}

// Create TestConfig to the test using the functions created in the meshnet library.
func CreateTestConfig(port int) *mesh.TestConfig {
	testConfig := &mesh.TestConfig{}
	testConfig.Node = NewNodeConfig()
	testConfig.Reliable = transport.CreateReliable(testConfig.Node.PubKey)
	testConfig.Udp = transport.CreateUdp(port, "127.0.0.1")

	return testConfig
}

func CreateNode(config mesh.TestConfig) *mesh.Node {
	node, createNodeError := mesh.NewNode(config.Node)
	if createNodeError != nil {
		panic(createNodeError)
	}

	return node
}

// Create public key
func CreatePubKey() cipher.PubKey {
	pub, _ := cipher.GenerateKeyPair()
	return pub
}

// Create new node config
func NewNodeConfig() domain.NodeConfig {
	nodeConfig := domain.NodeConfig{}
	nodeConfig.PubKey = CreatePubKey()
	//nodeConfig.ChaCha20Key = CreateChaCha20Key()
	nodeConfig.MaximumForwardingDuration = 1 * time.Minute
	nodeConfig.RefreshRouteDuration = 5 * time.Minute
	nodeConfig.ExpireMessagesInterval = 5 * time.Minute
	nodeConfig.ExpireRoutesInterval = 5 * time.Minute
	nodeConfig.TimeToAssembleMessage = 5 * time.Minute
	nodeConfig.TransportMessageChannelLength = 100

	return nodeConfig
}

// Create node list
func (self *NodeManager) CreateNodeConfigList(n int) {
	self.ConfigList = make(map[cipher.PubKey]*mesh.TestConfig)
	self.NodesList = make(map[cipher.PubKey]*mesh.Node)
	if self.Port == 0 {
		self.Port = 10000
	}
	for a := 1; a <= n; a++ {
		self.AddNode()
	}
}

// Add Node to Nodes List
func (self *NodeManager) AddNode() int {
	if len(self.ConfigList) == 0 {
		self.ConfigList = make(map[cipher.PubKey]*mesh.TestConfig)
		self.NodesList = make(map[cipher.PubKey]*mesh.Node)
	}
	config := CreateTestConfig(self.Port)
	self.ConfigList[config.Node.PubKey] = config
	self.Port++
	node := CreateNode(*config)
	self.NodesList[config.Node.PubKey] = node
	self.PubKeyList = append(self.PubKeyList, config.Node.PubKey)
	index := len(self.NodesList) - 1
	return index
}

// Connect the node list
func (self *NodeManager) ConnectNodes() {

	var index2, index3 int
	var lenght int = len(self.ConfigList)

	if lenght > 1 {
		for index1, pubKey1 := range self.PubKeyList {

			if index1 == 0 {
				index2 = 1
			} else {
				if index1 == lenght-1 {
					index2 = index1 - 1
					index3 = 0
				} else {
					index2 = index1 - 1
					index3 = index1 + 1
				}
			}
			pubKey2 := self.PubKeyList[index2]
			config1 := self.ConfigList[pubKey1]
			config2 := self.ConfigList[pubKey2]
			ConnectNodeToNode(config1, config2)

			if index3 > 0 {
				pubKey3 := self.PubKeyList[index3]
				config3 := self.ConfigList[pubKey3]
				ConnectNodeToNode(config1, config3)
			}
			self.NodesList[pubKey1].AddTransportToNode(*config1)
		}
	}
}

// Connect Node1 (config1) to Node2 (config2)
func ConnectNodeToNode(config1, config2 *mesh.TestConfig) {
	var addr bytes.Buffer
	addr.WriteString(config2.Udp.ExternalAddress)
	addr.WriteString(":")
	addr.WriteString(strconv.Itoa(int(config2.Udp.ListenPortMin)))
	config1.AddPeerToConnect(addr.String(), config2)
	addr.Reset()
}

// Add a new transport to node
func (self *NodeManager) AddTransportsToNode(config mesh.TestConfig, indexNode int) {
	nodePubKey := self.PubKeyList[indexNode]
	node := self.NodesList[nodePubKey]
	node.AddTransportToNode(config)
}

// Get all transports from one node
func (self *NodeManager) GetTransportsFromNode(indexNode int) []transport.Transport {
	nodePubKey := self.PubKeyList[indexNode]
	node := self.NodesList[nodePubKey]
	return node.GetTransports()
}

func (self *NodeManager) RemoveTransportsFromNode(indexNode int, transport transport.Transport) {
	nodePubKey := self.PubKeyList[indexNode]
	node := self.NodesList[nodePubKey]
	node.RemoveTransport(transport)
}

// Obtain port for to use in the creating from node
func (self *NodeManager) GetPort() int {
	port := self.Port
	self.Port++
	return port
}

// Connect node to netwotk
func (self *NodeManager) ConnectNodeToNetwork() (int, int) {
	// Create new node
	index1 := self.AddNode()
	index2 := self.ConnectNodeRandomly(index1)
	return index1, index2
}

// Connect Node Randomly
func (self *NodeManager) ConnectNodeRandomly(index1 int) int {
	var index2, rang int
	rang = len(self.ConfigList)
	for i := 0; i < 3; i++ {
		rand.Seed(time.Now().UTC().UnixNano())
		index2 = rand.Intn(rang)
		if index2 == index1 && i == 2 {
			fmt.Fprintf(os.Stderr, "Error Node %v not connected\n", index1)
			index2 = -1
			break
		} else if index2 != index1 {
			fmt.Fprintf(os.Stdout, "Connect node %v to node %v and vice versa.\n", index1, index2)
			pubKey1 := self.PubKeyList[index1]
			config1 := self.ConfigList[pubKey1]
			pubKey2 := self.PubKeyList[index2]
			config2 := self.ConfigList[pubKey2]
			ConnectNodeToNode(config1, config2)
			ConnectNodeToNode(config2, config1)
			break
		}
	}
	return index2
}

// Create routes from a node
func (self *NodeManager) BuildRoutes() {
	self.Routes = make(map[RouteKey]Route)
	for _, pubKey := range self.PubKeyList {
		self.FindRoute(pubKey)
	}
}
