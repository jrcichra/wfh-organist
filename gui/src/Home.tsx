import { useEffect, useRef, useState } from "react";
import Display from "./components/Display";
import Panic from "./components/Panic";
import Piston from "./components/Piston";
import RockerTab from "./components/RockerTab";
import "./Home.css";
import Play from "./components/Play";
import Stop from "./components/Stop";
import Set from "./components/Set";
import "./components/Video.css";

//@ts-ignore
import { Piano, KeyboardShortcuts, MidiNumbers } from "react-piano";
import "react-piano/dist/styles.css";

type StopType = {
  name: string;
  group: string;
  pressed: boolean;
};

function Home() {
  const [selectedPiston, setSelectedPiston] = useState<string>("-");
  const [pressedPiston, setPressedPiston] = useState<boolean>(false);

  const [midiFile, setMidiFile] = useState<string>("");
  const [midiFiles, setMidiFiles] = useState<string[]>([]);

  const [stops, setStops] = useState<StopType[]>([]);
  const [setMode, setSetMode] = useState<string>("false");

  const [pianoChannel, setPianoChannel] = useState<number>(1);

  const [midiLog, setMidiLog] = useState<string>("");

  const websocket = useRef<WebSocket>();

  const firstNote = MidiNumbers.fromNote("c3");
  const lastNote = MidiNumbers.fromNote("f5");
  const keyboardShortcuts = KeyboardShortcuts.create({
    firstNote: firstNote,
    lastNote: lastNote,
    keyboardConfig: KeyboardShortcuts.HOME_ROW,
  });

  var lastGroup: string;

  function setPressed(id: string, pressed: boolean) {
    if (selectedPiston !== "0") {
      setSelectedPiston("0");
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
    setSelectedPiston(id.toString());
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
    // set up the websocket
    if (!websocket.current) {
      let wsProtoco = "";
      if (location.protocol === "https:") {
        wsProtoco = "wss";
      } else {
        wsProtoco = "ws";
      }
      websocket.current = new WebSocket(
        `${wsProtoco}://${document.location.host}/ws`
      );
      websocket.current.onopen = () => {
        console.log("Successfully Connected");
      };
      websocket.current.onclose = (event) => {
        console.log("Socket Closed Connection: ", event);
      };
      websocket.current.onerror = (error) => {
        console.log("Socket Error: ", error);
      };
      websocket.current.onmessage = (event) => {
        console.log("Socket Message: ", event.data);
        setMidiLog(`${midiLog}\n${event.data}\nbob`);
      };
    }

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
        <Play midiFile={midiFile} />
        <Stop />
        <select
          name="midiFile"
          id="midiFile"
          onChange={(e) => setMidiFile(e.currentTarget.value)}
        >
          <option value=""></option>
          {midiFiles.map((file: string) => (
            <>
              <option value={file}>{file}</option>
              <label htmlFor={file}>{file}</label>
            </>
          ))}
        </select>
        <p className="title">General Pistons</p>
        <div className="col">
          <Set text="Set" value="true" set={setSetMode} />
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
          src={
            import.meta.env.VITE_VIDEO_URL ?? "https://wfho-video.jrcichra.dev/"
          }
          alt="wfho-video"
          className="remoteVideo"
        />
        <select
          name="pianoChannel"
          id="pianoChannel"
          value={pianoChannel}
          onChange={(e) => setPianoChannel(Number(e.currentTarget.value))}
        >
          {[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16].map(
            (channel: number) => (
              <option key={channel} value={channel}>
                {channel}
              </option>
            )
          )}
        </select>
        <Piano
          noteRange={{ first: firstNote, last: lastNote }}
          playNote={(midiNumber: any) => {
            console.log(midiNumber);
            if (websocket.current && websocket.current.readyState === 1) {
              websocket.current.send(
                JSON.stringify({
                  type: "noteOn",
                  key: midiNumber,
                  velocity: 127,
                  channel: pianoChannel - 1,
                })
              );
            }
          }}
          stopNote={(midiNumber: any) => {
            console.log(midiNumber);
            if (websocket.current && websocket.current.readyState === 1) {
              websocket.current.send(
                JSON.stringify({
                  type: "noteOff",
                  key: midiNumber,
                  channel: pianoChannel - 1,
                })
              );
            }
          }}
          width={1000}
          keyboardShortcuts={keyboardShortcuts}
        />
        <textarea id="midilog" value={midiLog}></textarea>
      </div>
    </div>
  );
}

export default Home;
