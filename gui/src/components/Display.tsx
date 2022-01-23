import './Display.css';

function Display({ value }: { value: string }) {

    return (
        <button className="display">{value}</button>
    )
};

export default Display;
