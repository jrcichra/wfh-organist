import { useEffect, useRef, useState } from 'react';
import './RockerTab.css';

function RockerTab({ text, on, off }: { text: string, on?: string, off?: string }) {

    const [pressed, setPressed]: [boolean, any] = useState(false);
    const [className, setClassName]: [string, any] = useState('button');

    const isMounted = useRef(false);

    useEffect(() => {
        (async () => {
            if (isMounted.current) {
                if (pressed) {
                    if (on !== undefined) {
                        setClassName('buttonActive');
                        // fetch post
                        fetch('/api/midi/raw', {
                            method: 'POST',
                            headers: {
                                'Content-Type': 'text/plain'
                            },
                            body: on,
                        });
                    }
                } else {
                    if (off !== undefined) {
                        setClassName('button');
                        // fetch post
                        fetch('/api/midi/raw', {
                            method: 'POST',
                            headers: {
                                'Content-Type': 'text/plain'
                            },
                            body: off,
                        });
                    }
                }
            } else {
                isMounted.current = true;
            }
        })()
    }, [pressed]);

    return (
        <button onClick={() => setPressed(!pressed)} className={className}>{text}</button>
    )
};

export default RockerTab;
