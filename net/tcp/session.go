package tcp

import (
	"net"
	"sync/atomic"
	"libs/log"
	"sync"
	"errors"
)

var (
	Error_Session_Closed = errors.New("session closed")
	Error_Session_Send_Chan_Full = errors.New("session send chan full")
)

var global_sessionId uint64

type Session struct {
	msgParser *MsgParser
	codec Codec
	sessionId uint64
	conn net.Conn

	errChan chan interface{}
	sendChan chan interface{}
	closeChan chan struct{}
	closeFlag bool
	lock sync.RWMutex

	data interface{}
}

func (session *Session) Id() uint64 {
	return session.sessionId
}

func (session *Session) ErrChan() <-chan interface{}{
	return  session.errChan
}

func (session *Session) Send(msg ...interface{}) error {
	session.lock.RLock()
	if session.closeFlag {
		session.lock.RUnlock()
		return Error_Session_Closed
	}

	select {
	case session.sendChan <- msg:
	default:
		session.lock.RUnlock()
		log.Error("session send chan full cap:%d", cap(session.sendChan))
		session.Close()
		return Error_Session_Send_Chan_Full
	}
	session.lock.RUnlock()

	return nil
}

func (session *Session) Receive() (interface{}, error) {
	data, err := session.msgParser.Read(session.conn)
	if err != nil {
		return nil, err
	}

	return session.codec.Unmarshal(data)
}


func (session *Session) Close() {
	session.lock.Lock()
	defer session.lock.Unlock()
	if session.closeFlag {
		return
	}

	session.closeFlag = true
	session.conn.Close()
	close(session.closeChan)
	close(session.sendChan)

	//获取未读取到的数据
	if csc, ok := session.codec.(CloseSendChan); ok {
		csc.CloseEnd(session.sendChan)
	}
	close(session.errChan)
}

func (session *Session) sendLoop() {
	for {
		select {
		case <- session.closeChan:
			goto __END
		case msg, ok := <- session.sendChan:
			if !ok {
				goto __END
			}
			data, err := session.codec.Marshal(msg)
			if err != nil {
				log.Error("session sendLoop codec.Marshal fail:%v", err)
				session.sendErr2Chan(err)
				continue
			}

			err = session.msgParser.Write(session.conn, data)
			if err != nil {
				log.Error("session sendLoop msgParser.Write fail:%v", err)
				session.sendErr2Chan(err)
				continue
			}
		}
	}

__END:
	session.Close()
}

func (session *Session)sendErr2Chan(err error){
	select {
	// 框架使用者没有从errChan中读取数据，导致buffer使用完,此处逻辑阻塞
	// 此逻辑最开始没有，本次新加后旧服务未必能及时更改代码，处理err，一样导致此处逻辑阻塞
	case session.errChan <- err:
		log.Debug("send err to error channl")
	default:
		// errChan满时执行default,防止框架逻辑阻塞
	}
}

func (session *Session) RemoteAddr() net.Addr {
	return session.conn.RemoteAddr()
}

func (session *Session) Data() interface{} {
	return session.data
}

func newSession(conn net.Conn, msgParser *MsgParser, protocol Protocol, sendChanSize int) *Session {
	var errChanSize = 10
	if sendChanSize > 0{
		errChanSize = sendChanSize
	}
	sessionId := atomic.AddUint64(&global_sessionId, 1)
	session := &Session{
		conn: conn,
		sessionId: sessionId,
		msgParser: msgParser,
		codec: protocol.NewCodec(),
		sendChan: make(chan interface{}, sendChanSize),
		closeChan: make(chan struct{}),
		errChan:make(chan interface{}, errChanSize),
	}

	//开启写协程
	go session.sendLoop()

	return session
}