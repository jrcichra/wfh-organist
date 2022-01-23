import Peer from 'peerjs';
import { useEffect, useRef, useState } from 'react';
import { useLocation } from 'react-router-dom';
import Display from './components/Display';
import Piston from './components/Piston';
import RockerTab from './components/RockerTab';
import './Home.css';
import Video from './Video';


function Home() {

  const location = useLocation();
  const [myID, setMyID]: [any, any] = useState('');
  const [remoteID, setRemoteID]: [any, any] = useState('');
  const peer: any = useRef(null);

  const [selectedPiston, setSelectedPiston]: [any, any] = useState(null);

  const [localStream, setLocalStream]: [any, any] = useState(null);
  const [remoteStream, setRemoteStream]: [any, any] = useState(null);

  useEffect(() => {

    if (location.pathname === '/server') {
      peer.current = new Peer('wfh-organist-server');
      setMyID('wfh-organist-server');
      setRemoteID('wfh-organist-client');
    } else {
      peer.current = new Peer('wfh-organist-client');
      setMyID('wfh-organist-client');
      setRemoteID('wfh-organist-server');
    }

    peer.current.on('open', (id: any) => {
      console.log('My peer ID is: ' + id);
      setMyID(id);
    });

    peer.current.on('connection', (conn: any) => {
      console.log('Connection made');
      conn.on('data', (data: any) => {
        console.log('Received data: ' + data);
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

  useEffect(() => {
    if (myID == 'wfh-organist-client') {
      videoCall();
    }
  }, [myID]);

  return (
    <div className="wrapper">
      {/* <div className="title">
      </div> */}
      {/* <div className="col">
        <p>My ID: {myID}</p>
      </div> */}
      <p className="title">Swell Organ</p>
      <div className="col">
        <RockerTab text="Bourdon 16'" />
        <RockerTab text="Gedackt 8'" />
        <RockerTab text="Viola 8'" />
        <RockerTab text="Viola Celeste 8'" />
        <RockerTab text="Spitz prinzipal 4'" />
        <RockerTab text="Koppel flote 4'" />
        <RockerTab text="Nasat 2-2/3'" />
        <RockerTab text="Blockflote 2'" />
        <RockerTab text="Basson 16'" />
        <RockerTab text="Trompette 8'" />
        <RockerTab text="Tremulant" />
        <RockerTab text="MIDI to Swell" />
      </div>
      <p className="title">Great Organ</p>
      <div className="col">
        <RockerTab text="Principal 8'" />
        <RockerTab text="Gedackt 8'" />
        <RockerTab text="Koppel flote 4'" />
        <RockerTab text="Super Octave 2'" />
        <RockerTab text="Mixture IV" />
        <RockerTab text="Chimes" />
        <RockerTab text="Tremulant" />
        <RockerTab text="Swell to Great" />
        <RockerTab text="MIDI to Great" />
      </div>
      <p className="title">Pedal Organ</p>
      <div className="col">
        <RockerTab text="Bourdon 16'" />
        <RockerTab text="Lieb lich gedackt 16'" />
        <RockerTab text="Octave 8'" />
        <RockerTab text="Gedackt 8'" />
        <RockerTab text="Choral bass 4'" />
        <RockerTab text="Mixture II" />
        <RockerTab text="Basson 16'" />
        <RockerTab text="Trompette 8'" />
        <RockerTab text="Great to Pedal" />
        <RockerTab text="Swell to Pedal" />
        <RockerTab text="MIDI to Pedal" />
      </div>
      <p className="title">General</p>
      <div className="col">
        <RockerTab text="Memory B" />
        <RockerTab text="Add Stops" />
        <RockerTab text="Bass Coupler" />
        <RockerTab text="Melody Coupler" />
        <RockerTab text="Romantic Tuning Off" />
        <RockerTab text="Reverb" />
        <RockerTab text="." />
        <RockerTab text="Console Speakers Off" />
        <RockerTab text="External Speakers Off" />
      </div>
      <p className="title">General Pistons</p>
      <div className="col">
        <Piston text="Set" />
        <span className="pistonGap"></span>
        <Piston text="1" value="1" set={setSelectedPiston} />
        <Piston text="2" value="2" set={setSelectedPiston} />
        <Piston text="3" value="3" set={setSelectedPiston} />
        <Piston text="4" value="4" set={setSelectedPiston} />
        <Piston text="5" value="5" set={setSelectedPiston} />
        <Piston text="6" value="6" set={setSelectedPiston} />
        <Piston text="7" value="7" set={setSelectedPiston} />
        <span className="pistonGap"></span>
        <Piston text="Cancel" value="-" set={setSelectedPiston} />
        <span className="pistonGap"></span>
        <Display value={selectedPiston} />
      </div>

      <div className="col">
        <div className="videos">
          <div>
            <Video className="localVideo" muted autoPlay playsInline srcObject={localStream} />
          </div>
          <div>
            <Video className="remoteVideo" autoPlay playsInline srcObject={remoteStream} />
          </div>
        </div>
      </div>
    </div>
  );
}

export default Home;
