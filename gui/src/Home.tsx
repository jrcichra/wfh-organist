import { useEffect, useRef, useState } from "react";
import { useLocation } from "react-router-dom";
import Display from "./components/Display";
import Panic from "./components/Panic";
import Piston from "./components/Piston";
import RockerTab from "./components/RockerTab";
import "./Home.css";
import Play from "./components/Play";
import Stop from "./components/Stop";
import "./components/Video.css";

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
  const [selectedPiston, setSelectedPiston]: [any, any] = useState("-");
  const [pressedPiston, setPressedPiston]: [any, any] = useState(false);

  const [midiFile, setMidiFile]: [any, any] = useState("");
  const [midiFiles, setMidiFiles]: [any, any] = useState([]);

  const [stops, setStops]: [any, any] = useState([]);
  const [setMode, setSetMode]: [any, any] = useState("false");

  var lastGroup: string;

  function setPressed(id: string, pressed: boolean) {
    if (selectedPiston !== 0) {
      setSelectedPiston(0);
    }
    let tempStops: StopType[] = [...stops];

    tempStops.forEach((stop: StopType) => {
      if (`stop/${stop.group}/${stop.name}` === id) {
        stop.pressed = pressed;
      }
    });

    setStops(tempStops);
  }

  function setPiston(id: number) {
    setSelectedPiston(id);
    setPressedPiston(true);
    if (id === 0) {
      // this is the cancel button. All stops should be raised
      let tempStops: StopType[] = [...stops];
      tempStops.forEach((stop: StopType) => {
        stop.pressed = false;
      });
      // update the stops visually and on the backend
      setStops(tempStops);
      fetch("/api/midi/stops", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(stops),
      });
    }
  }

  useEffect(() => {
    if (pressedPiston) {
      if (setMode === "true") {
        setSetMode("false");
        // store the stops under the value of selectedPiston
        fetch(`/api/midi/stops?piston=${selectedPiston}`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify(stops),
        });
      } else {
        // ask to apply the state for this piston
        fetch(`/api/midi/piston`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: String(selectedPiston),
        })
          .then((res) => res.json())
          .then((data) => {
            setStops(data);
          });
      }
      setPressedPiston(false);
    }
  }, [selectedPiston, pressedPiston]);

  useEffect(() => {
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
                            id={`stop/${stop.group}/${stop.name}`}
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
          <Piston text="1" value="1" set={setPiston} />
          <Piston text="2" value="2" set={setPiston} />
          <Piston text="3" value="3" set={setPiston} />
          <Piston text="4" value="4" set={setPiston} />
          <Piston text="5" value="5" set={setPiston} />
          <Piston text="6" value="6" set={setPiston} />
          <Piston text="7" value="7" set={setPiston} />
          <span className="pistonGap"></span>
          <Piston text="Cancel" value="0" set={setPiston} />
          <span className="pistonGap"></span>
          <Display value={selectedPiston} />
        </div>
      </div>
      <div className="col">
        <img
          src="https://wfho-video.jrcichra.dev/"
          alt="wfho-video"
          className="remoteVideo"
        />
      </div>
    </div>
  );
}

export default Home;
