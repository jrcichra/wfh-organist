import { useEffect, useRef, useState } from "react";
import "./RockerTab.css";

function RockerTab({
  text,
  id,
  pressed,
  setPressed,
}: {
  text: string;
  id: number;
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
          body: String(id),
        });
      } else {
        isMounted.current = true;
      }
    })();
  }, [pressed]);

  return (
    <button onClick={() => setPressed(id, !pressed)} className={className} >
      {text}
    </button >
  );
}

export default RockerTab;
