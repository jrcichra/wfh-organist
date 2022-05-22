import "./Panic.css";

function Panic() {
  return (
    <button
      onClick={() => {
        fetch("/api/midi/panic");
      }}
      className="panicButton"
    >
      PANIC
    </button>
  );
}

export default Panic;
