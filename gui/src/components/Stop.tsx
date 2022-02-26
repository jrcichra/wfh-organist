import './Stop.css';

function Stop() {

    function sendStop() {
        // fetch stop
        fetch('/api/midi/file/stop');
    }

    return (
        <button onClick={() => { sendStop() }} className="stopButton">STOP</button>
    )
};

export default Stop;
