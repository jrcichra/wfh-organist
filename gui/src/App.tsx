import Peer from 'peerjs';
import { useEffect, useRef, useState } from 'react';
import './App.css';
import Video from './Video';


function App() {

  const [myID, setMyID]: [any, any] = useState('');
  const [remoteID, setRemoteID]: [any, any] = useState('');
  const peer: any = useRef(null);
  const [message, setMessage]: [any, any] = useState('');
  const [messages, setMessages]: [any, any] = useState([]);

  const [localStream, setLocalStream]: [any, any] = useState(null);
  const [remoteStream, setRemoteStream]: [any, any] = useState(null);

  useEffect(() => {
    peer.current = new Peer();

    peer.current.on('open', (id: any) => {
      console.log('My peer ID is: ' + id);
      setMyID(id);
    });

    peer.current.on('connection', (conn: any) => {
      console.log('Connection made');
      conn.on('data', (data: any) => {
        console.log('Received data: ' + data);
        setMessages((messages: any) => [...messages, data]);
      });
    });

    peer.current.on('call', (call: any) => {
      console.log('Received call');
      (async () => {
        const stream = await navigator.mediaDevices.getUserMedia({ video: true, audio: true });
        setLocalStream(stream);
        call.answer(stream);
        call.on('stream', (remoteStream: any) => {
          console.log('Received remote stream');
          setRemoteStream(remoteStream);
        });
      })();
    });

  }, []);

  const send = () => {
    console.log('Sending message: ' + message);
    const conn = peer.current.connect(remoteID);

    conn.on('open', () => {
      const msgObj = {
        sender: myID,
        message: message
      };
      conn.send(msgObj);
      setMessages([...messages, msgObj]);
      setMessage('');
    });
  };

  const videoCall = () => {
    (async () => {
      console.log('Starting call');
      const stream = await navigator.mediaDevices.getUserMedia({ video: true, audio: true });
      setLocalStream(stream);
      const call = peer.current.call(remoteID, stream);
      call.on('stream', (remoteStream: any) => {
        setRemoteStream(remoteStream);
      });
    })();
  };

  return (
    <div className="wrapper">
      <div className="col">
        <h1>My ID: {myID}</h1>

        <label>Friend ID:</label>
        <input
          type="text"
          value={remoteID}
          onChange={e => { setRemoteID(e.target.value); }} />

        <br />
        <br />

        <label>Message:</label>
        <input
          type="text"
          value={message}
          onChange={e => { setMessage(e.target.value); }} />
        <button onClick={send}>Send</button>

        <button onClick={videoCall}>Video Call</button>
        {
          messages.map((message: any, i: any) => {
            return (
              <div key={i}>
                <h3>{message.sender}:</h3>
                <p>{message.message}</p>
              </div>

            )
          })
        }
      </div>

      <div className="col">
        <div>
          <Video id="localVideo" muted autoPlay playsInline srcObject={localStream} />
        </div>
        <div>
          <Video id="remoteVideo" autoPlay playsInline srcObject={remoteStream} />
        </div>
      </div>

    </div>
  );
}

export default App;
