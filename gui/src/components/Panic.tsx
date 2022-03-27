import "./Panic.css";

function Panic() {
  return (
    <button
      onClick={() => {
        fetch("/api/midi/raw");
      }}
      className="panicButton"
    >
      PANIC
    </button>
  );
}

export default Panic;
