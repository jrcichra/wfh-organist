import './Play.css';

function Play({ midiFile }: { midiFile: string }) {

    function sendPlay() {
        // fetch post
        fetch('/api/midi/file/play', {
            method: 'POST',
            headers: {
                'Content-Type': 'text/plain'
            },
            body: midiFile,
        });
    }

    return (
        <button onClick={() => { sendPlay() }} className="playButton">PLAY</button>
    )
};

export default Play;
