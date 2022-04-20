import { useEffect, useState } from "react";
import "./Set.css";

function Set({ text, value, set }: { text: string; value: string; set?: any }) {
  const [className, setClassName]: [string, any] = useState("button");
  const [pressed, setPressed]: [boolean, any] = useState(false);

  useEffect(() => {
    if (pressed) {
      setClassName("setButtonActive");
    } else {
      setClassName("setButton");
    }
  }, [pressed]);

  return (
    <button
      onMouseDown={() => {
        set(parseInt(value) == Number(value) ? Number(value) : value);
        setPressed(!pressed);
      }}
      className={className}
    >
      {text}
    </button>
  );
}

export default Set;
