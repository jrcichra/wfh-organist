import { useEffect, useState } from 'react';
import './RockerTab.css';

function RockerTab({ text }: { text: string }) {

    const [pressed, setPressed]: [boolean, any] = useState(false);

    const [className, setClassName]: [string, any] = useState('button');

    useEffect(() => {
        if (pressed) {
            setClassName('buttonActive');
        } else {
            setClassName('button');
        }
    }, [pressed]);

    return (
        <button onClick={() => setPressed(!pressed)} className={className}>{text}</button>
    )
};

export default RockerTab;
