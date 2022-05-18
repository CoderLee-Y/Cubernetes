package heartbeat

import (
	cubeconfig "Cubernetes/config"
	"Cubernetes/pkg/apiserver/watchobj"
	"Cubernetes/pkg/object"
	"bufio"
	"encoding/json"
	"log"
	"net"
	"strconv"
	"sync"
	"time"
)

const MSG_DELIM byte = 26
const MSG_HEARTBEAT = "alive"

const INTERVAL = 5 * time.Second
const TIMEOUT = 15 * time.Second

var conn net.Conn
var connected bool
var node object.Node
var lastUpdate time.Time

var timeLock sync.Mutex
var connLock sync.Mutex

// CheckConn
// package heartbeat is designed for cubelet
func CheckConn() bool {
	connLock.Lock()
	ret := connected
	connLock.Unlock()
	return ret
}

func closeConn() {
	connLock.Lock()
	if connected {
		_ = conn.Close()
		connected = false
		log.Println("Heartbeat connection closed, stop all watching")
		watchobj.StopAll()
	}
	connLock.Unlock()
}

func updateTime() {
	timeLock.Lock()
	lastUpdate = time.Now()
	timeLock.Unlock()
}

func getTime() time.Time {
	timeLock.Lock()
	t := lastUpdate
	timeLock.Unlock()
	return t
}

func maintainHealth() {
	defer closeConn()

	for {
		if time.Since(getTime()) > TIMEOUT {
			log.Println("Timeout, close conn")
			return
		}

		buf, err := json.Marshal(node)
		if err != nil {
			log.Println("Fail to marshal Node, err: ", err)
			return
		}
		buf = append(buf, MSG_DELIM)
		_, err = conn.Write(buf)
		if err != nil {
			log.Println("Fail to send Node message, err: ", err)
			return
		}
		time.Sleep(INTERVAL)
	}
}

func updateHeartBeat() {
	var err error
	conn, err = net.Dial("tcp", cubeconfig.APIServerIp+":"+strconv.Itoa(cubeconfig.HeartbeatPort))
	if err != nil {
		log.Fatal("Fail to dial heartbeat server", err)
		return
	}

	lastUpdate = time.Now()
	connected = true
	defer closeConn()

	go maintainHealth()

	reader := bufio.NewReader(conn)
	for {
		_, err := reader.ReadBytes(MSG_DELIM)
		if err != nil {
			log.Println("Fail to read from conn")
			return
		}
		updateTime()
	}
}

// InitNode
// package heartbeat is designed for cubelet
func InitNode(n object.Node) {
	node = n
	node.Status.Condition.Ready = true
	connected = false
	go updateHeartBeat()
}
