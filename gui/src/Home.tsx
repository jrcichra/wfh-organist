import Peer from 'peerjs';
import { useEffect, useRef, useState } from 'react';
import { useLocation } from 'react-router-dom';
import Display from './components/Display';
import Panic from './components/Panic';
import Piston from './components/Piston';
import RockerTab from './components/RockerTab';
import './Home.css';
import Video from './components/Video';

const audioOptions: MediaTrackConstraints = {
  autoGainControl: false,
  channelCount: 2,
  echoCancellation: false,
  latency: 0,
  noiseSuppression: false,
  sampleRate: 16000,
  sampleSize: 8,
}

const videoOptions: MediaTrackConstraints = {
  frameRate: 2,
  width: 640,
  height: 480,
}

function Home() {

  const location = useLocation();
  const [myID, setMyID]: [any, any] = useState('');
  const [remoteID, setRemoteID]: [any, any] = useState('');
  const peer: any = useRef(null);

  const [selectedPiston, setSelectedPiston]: [any, any] = useState(null);

  const [localStream, setLocalStream]: [any, any] = useState(null);
  const [remoteStream, setRemoteStream]: [any, any] = useState(null);

  useEffect(() => {

    if (new URLSearchParams(location.search).get("mode") === 'server') {
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
        const stream = await navigator.mediaDevices.getUserMedia({
          video: videoOptions, audio: audioOptions,
        });
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
      const stream = await navigator.mediaDevices.getUserMedia({ video: videoOptions, audio: audioOptions });
      setLocalStream(stream);
      const call = peer.current.call(remoteID, stream);
      call.on('stream', (remoteStream: any) => {
        setRemoteStream(remoteStream);
      });
    })();
  };

  useEffect(() => {
    if (myID === 'wfh-organist-client') {
      videoCall();
    }
  }, [myID]);

  return (
    <div className="wrapper">
      <div className="stop-container">
        <p className="title">Swell Organ</p>
        <div className="col">
          <RockerTab text="Bourdon 16'" on="b0 63 00 b0 62 0b b0 06 7f" off="b0 63 00 b0 62 0b b0 06 00" />
          <RockerTab text="Gedackt 8'" on="b0 63 00 b0 62 28 b0 06 7f" off="b0 63 00 b0 62 28 b0 06 00" />
          <RockerTab text="Viola 8'" on="b0 63 00 b0 62 22 b0 06 7f" off="b0 63 00 b0 62 22 b0 06 00" />
          <RockerTab text="Viola Celeste 8'" on="b0 63 00 b0 62 23 b0 06 7f" off="b0 63 00 b0 62 23 b0 06 00" />
          <RockerTab text="Spitz prinzipal 4'" on="b0 63 00 b0 62 38 b0 06 7f" off="b0 63 00 b0 62 38 b0 06 00" />
          <RockerTab text="Koppel flote 4'" on="b0 63 00 b0 62 3f b0 06 7f" off="b0 63 00 b0 62 3f b0 06 00" />
          <RockerTab text="Nasat 2-2/3'" on="b0 63 00 b0 62 4c b0 06 7f" off="b0 63 00 b0 62 4c b0 06 00" />
          <RockerTab text="Blockflote 2'" on="b0 63 00 b0 62 54 b0 06 7f" off="b0 63 00 b0 62 54 b0 06 00" />
          <RockerTab text="Basson 16'" on="b0 63 00 b0 62 72 b0 06 7f" off="b0 63 00 b0 62 72 b0 06 00" />
          <RockerTab text="Trompette 8'" on="b0 63 01 b0 62 00 b0 06 7f" off="b0 63 01 b0 62 00 b0 06 00" />
          <RockerTab text="Tremulant" on="b0 63 01 b0 62 30 b0 06 7f" off="b0 63 01 b0 62 30 b0 06 00" />
          <RockerTab text="MIDI to Swell" on="b0 63 01 b0 62 5e b0 06 7f" off="b0 63 01 b0 62 5e b0 06 00" />
        </div>
        <p className="title">Great Organ</p>
        <div className="col">
          <RockerTab text="Principal 8'" on="b1 63 00 b1 62 1f b1 06 7f" off="b1 63 00 b1 62 1f b1 06 00" />
          <RockerTab text="Gedackt 8'" on="b1 63 00 b1 62 28 b1 06 7f" off="b1 63 00 b1 62 28 b1 06 00" />
          <RockerTab text="Octave 4'" on="b1 63 00 b1 62 38 b1 06 7f" off="b1 63 00 b1 62 38 b1 06 00" />
          <RockerTab text="Koppel flote 4'" on="b1 63 00 b1 62 3f b1 06 7f" off="b1 63 00 b1 62 3f b1 06 00" />
          <RockerTab text="Super Octave 2'" on="b1 63 00 b1 62 50 b1 06 7f" off="b1 63 00 b1 62 50 b1 06 00" />
          <RockerTab text="Mixture IV" on="b1 63 00 b1 62 64 b1 06 7f" off="b1 63 00 b1 62 64 b1 06 00" />
          <RockerTab text="Chimes" on="b1 63 01 b1 62 21 b1 06 7f" off="b1 63 01 b1 62 21 b1 06 00" />
          <RockerTab text="Tremulant" on="b1 63 01 b1 62 30 b1 06 7f" off="b1 63 01 b1 62 30 b1 06 00" />
          <RockerTab text="Swell to Great" on="b1 63 01 b1 62 77 b1 06 7f" off="b1 63 01 b1 62 77 b1 06 00" />
          <RockerTab text="MIDI to Great" on="b1 63 01 b1 62 5f b1 06 7f" off="b1 63 01 b1 62 5f b1 06 00" />
          <span className="pistonGap"></span>
          <span className="pistonGap"></span>
          <span className="pistonGap"></span>
          <Panic data="b0 7b 00 b1 7b 00 b2 7b 00" />
        </div>
        <p className="title">Pedal Organ</p>
        <div className="col">
          <RockerTab text="Bourdon 16'" on="b2 63 00 b2 62 0c b2 06 7f" off="b2 63 00 b2 62 0c b2 06 00" />
          <RockerTab text="Lieb lich gedackt 16'" on="b2 63 00 b2 62 0f b2 06 7f" off="b2 63 00 b2 62 0f b2 06 00" />
          <RockerTab text="Octave 8'" on="b2 63 00 b2 62 1f b2 06 7f" off="b2 63 00 b2 62 1f b2 06 00" />
          <RockerTab text="Gedackt 8'" on="b2 63 00 b2 62 28 b2 06 7f" off="b2 63 00 b2 62 28 b2 06 00" />
          <RockerTab text="Choral bass 4'" on="b2 63 00 b2 62 38 b2 06 7f" off="b2 63 00 b2 62 38 b2 06 00" />
          <RockerTab text="Mixture II" on="b2 63 00 b2 62 64 b2 06 7f" off="b2 63 00 b2 62 64 b2 06 00" />
          <RockerTab text="Basson 16'" on="b2 63 00 b2 62 72 b2 06 7f" off="b2 63 00 b2 62 72 b2 06 00" />
          <RockerTab text="Trompette 8'" on="b2 63 01 b2 62 00 b2 06 7f" off="b2 63 01 b2 62 00 b2 06 00" />
          <RockerTab text="Great to Pedal" on="b2 63 01 b2 62 78 b2 06 7f" off="b2 63 01 b2 62 78 b2 06 00" />
          <RockerTab text="Swell to Pedal" on="b2 63 01 b2 62 77 b2 06 7f" off="b2 63 01 b2 62 77 b2 06 00" />
          <RockerTab text="MIDI to Pedal" on="b2 63 01 b2 62 60 b2 06 7f" off="b2 63 01 b2 62 60 b2 06 00" />
        </div>
        <p className="title">General</p>
        <div className="col">
          <RockerTab text="Memory B" />
          <RockerTab text="Add Stops" />
          <RockerTab text="Bass Coupler" on="b7 63 01 b7 62 44 b7 06 7f" off="b7 63 01 b7 62 44 b7 06 00" />
          <RockerTab text="Melody Coupler" on="b7 63 01 b7 62 45 b7 06 7f" off="b7 63 01 b7 62 45 b7 06 00" />
          <RockerTab text="Romantic Tuning Off" on="b7 63 01 b7 62 65 b7 06 7f" off="b7 63 01 b7 62 65 b7 06 00" />
          <RockerTab text="Reverb" />
          <RockerTab text="." on="b7 63 01 b7 62 31 b7 06 7f" off="b7 63 01 b7 62 31 b7 06 00" />
          <RockerTab text="Console Speakers Off" on="b7 63 01 b7 62 69 b7 06 7f" off="b7 63 01 b7 62 69 b7 06 00" />
          <RockerTab text="External Speakers Off" on="b7 63 01 b7 62 74 b7 06 7f" off="b7 63 01 b7 62 74 b7 06 00" />
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
      </div>
      <div className="col">
        <div className="videos">
          <div>
            <Video title="Local" className="localVideo" muted autoPlay playsInline srcObject={localStream} />
          </div>
          <div>
            <Video title="Remote" className="remoteVideo" autoPlay playsInline srcObject={remoteStream} />
          </div>
        </div>
      </div>
    </div>
  );
}

export default Home;
