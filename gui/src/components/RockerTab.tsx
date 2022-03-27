import { useEffect, useRef, useState } from "react";
import "./RockerTab.css";

function RockerTab({
  name,
  id,
  pressed,
  setPressed,
}: {
  name: string;
  id: string;
  pressed: boolean;
  setPressed: any;
}) {
  const [className, setClassName]: [string, any] = useState("button");
  const isMounted = useRef(false);

  useEffect(() => {
    if (pressed) {
      setClassName("buttonActive");
    } else {
      setClassName("button");
    }
  }, [pressed]);

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
    <button onClick={() => setPressed(id, !pressed)} className={className}>
      {name}
    </button>
  );
}

export default RockerTab;
