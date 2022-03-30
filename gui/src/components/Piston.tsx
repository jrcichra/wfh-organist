import "./Piston.css";

function Piston({
  text,
  value,
  set,
}: {
  text: string;
  value: string;
  set?: any;
}) {
  return (
    <button
      onMouseDown={() => {
        set(parseInt(value) == Number(value) ? Number(value) : value);
      }}
      className="pistonButton"
    >
      {text}
    </button>
  );
}

export default Piston;
