import { useEffect, useRef, useState } from "react";
import "./RockerTab.css";

function RockerTab({
  text,
  id,
  initalPressed,
}: {
  text: string;
  id: string;
  initalPressed: boolean;
}) {
  const [className, setClassName]: [string, any] = useState("button");
  const [pressed, setPressed]: [boolean, any] = useState(initalPressed);
  const isMounted = useRef(false);

  useEffect(() => {
    if (initalPressed) {
      setClassName("buttonActive");
    } else {
      setClassName("button");
    }
  }, [initalPressed]);

  useEffect(() => {
    (async () => {
      if (isMounted.current) {
        if (pressed) {
          setClassName("buttonActive");
        } else {
          setClassName("button");
        }
        fetch("/api/midi/pushstop", {
          method: "POST",
          headers: {
            "Content-Type": "text/plain",
          },
          body: id,
        });
      } else {
        isMounted.current = true;
      }
    })();
  }, [pressed]);

  return (
    <button onClick={() => setPressed(!pressed)} className={className}>
      {text}
    </button>
  );
}

export default RockerTab;
