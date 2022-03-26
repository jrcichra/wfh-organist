import './Piston.css';

function Piston({ text, value, set }: { text: string, value: string, set?: any }) {

    return (
        <button onMouseDown={() => { set(value) }} className="pistonButton">{text}</button>
    )
};

export default Piston;
