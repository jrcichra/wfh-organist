import "./Display.css";

function Display({ value }: { value: string }) {
  return <button className="display">{value == "0" ? "-" : value}</button>;
}

export default Display;
