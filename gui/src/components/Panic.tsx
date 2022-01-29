import './Panic.css';

function Panic({ data }: { data: string }) {

    function sendPanic() {
        // fetch post
        fetch('/api/midi/raw', {
            method: 'POST',
            headers: {
                'Content-Type': 'text/plain'
            },
            body: data,
        });
    }

    return (
        <button onClick={() => { sendPanic() }} className="panicButton">PANIC</button>
    )
};

export default Panic;
