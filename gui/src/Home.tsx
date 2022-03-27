import Peer from "peerjs";
import { useEffect, useRef, useState } from "react";
import { useLocation } from "react-router-dom";
import Display from "./components/Display";
import Panic from "./components/Panic";
import Piston from "./components/Piston";
import RockerTab from "./components/RockerTab";
import "./Home.css";
import Video from "./components/Video";
import Play from "./components/Play";
import Stop from "./components/Stop";

const videoOptions: MediaTrackConstraints = {
  frameRate: 2,
  width: 640,
  height: 480,
};

type StopType = {
  name: string;
  group: string;
  pressed: boolean;
};

function Home() {
  const location = useLocation();
  const [myID, setMyID]: [any, any] = useState("");
  const [remoteID, setRemoteID]: [any, any] = useState("");
  const peer: any = useRef(null);

  const [selectedPiston, setSelectedPiston]: [any, any] = useState(null);

  const [localStream, setLocalStream]: [any, any] = useState(null);
  const [remoteStream, setRemoteStream]: [any, any] = useState(null);

  const [midiFile, setMidiFile]: [any, any] = useState("");
  const [midiFiles, setMidiFiles]: [any, any] = useState([]);

  const [stops, setStops]: [any, any] = useState([]);
  const [setMode, setSetMode]: [any, any] = useState("false");

  var lastGroup: string;

  function setPressed(id: string, pressed: boolean) {
    let tempStops: StopType[] = [...stops];

    tempStops.forEach((stop: StopType) => {
      if (`${stop.group}/${stop.name}` === id) {
        stop.pressed = pressed;
      }
    });

    setStops(tempStops);
  }

  useEffect(() => {
    if (setMode === "true") {
      setSetMode("false");
      //store the stops under the value of selectedPiston
      fetch("/api/midi/stops", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(stops),
      });
    }
  }, [selectedPiston]);

  useEffect(() => {
    if (new URLSearchParams(location.search).get("mode") === "server") {
      peer.current = new Peer("wfh-organist-server");
      setMyID("wfh-organist-server");
      setRemoteID("wfh-organist-client");
    } else {
      peer.current = new Peer("wfh-organist-client");
      setMyID("wfh-organist-client");
      setRemoteID("wfh-organist-server");
    }

    peer.current.on("open", (id: any) => {
      console.log("My peer ID is: " + id);
      setMyID(id);
    });

    peer.current.on("connection", (conn: any) => {
      console.log("Connection made");
      conn.on("data", (data: any) => {
        console.log("Received data: " + data);
      });
    });

    peer.current.on("call", (call: any) => {
      console.log("Received call");
      (async () => {
        const stream = await navigator.mediaDevices.getUserMedia({
          video: videoOptions,
          audio: false,
        });
        setLocalStream(stream);
        call.answer(stream);
        call.on("stream", (remoteStream: any) => {
          console.log("Received remote stream");
          setRemoteStream(remoteStream);
        });
      })();
    });

    // get the list of midi files
    fetch("/api/midi/files")
      .then((res) => res.json())
      .then((data) => {
        setMidiFiles(data);
      });

    // get the stops
    fetch("/api/midi/stops")
      .then((res) => res.json())
      .then((data) => {
        setStops(data);
      });
  }, []);

  const videoCall = () => {
    (async () => {
      console.log("Starting call");
      const stream = await navigator.mediaDevices.getUserMedia({
        video: videoOptions,
        audio: false,
      });
      setLocalStream(stream);
      const call = peer.current.call(remoteID, stream);
      call.on("stream", (remoteStream: any) => {
        setRemoteStream(remoteStream);
      });
    })();
  };

  useEffect(() => {
    if (myID === "wfh-organist-client") {
      videoCall();
    }
  }, [myID]);

  return (
    <div className="wrapper">
      <div className="stop-container">
        {stops.map((stop: StopType) => {
          if (stop.group != lastGroup) {
            lastGroup = stop.group;
            return (
              <>
                <p className="title">{`${stop.group} Organ`}</p>
                <div className="col">
                  {stops.map((stop: StopType) => {
                    if (stop.group === lastGroup) {
                      return (
                        <>
                          <RockerTab
                            name={stop.name}
                            id={`${stop.group}/${stop.name}`}
                            pressed={stop.pressed}
                            setPressed={setPressed}
                          />
                        </>
                      );
                    }
                  })}
                </div>
              </>
            );
          }
        })}
        <br></br>
        <br></br>
        <span className="pistonGap"></span>
        <span className="pistonGap"></span>
        <span className="pistonGap"></span>
        <Panic />
        {midiFiles.map((file: string) => (
          <>
            <input
              id={file}
              type="radio"
              name="midiFile"
              value={file}
              onClick={(e) => setMidiFile(e.currentTarget.value)}
            />
            <label htmlFor={file}>{file}</label>
          </>
        ))}
        <Play midiFile={midiFile} />
        <Stop />
        <p className="title">General Pistons</p>
        <div className="col">
          <Piston text="Set" value="true" set={setSetMode} />
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
            <Video
              title="Local"
              className="localVideo"
              muted
              autoPlay
              playsInline
              srcObject={localStream}
            />
          </div>
          <div>
            <Video
              title="Remote"
              className="remoteVideo"
              autoPlay
              playsInline
              srcObject={remoteStream}
            />
          </div>
        </div>
      </div>
    </div>
  );
}

export default Home;
