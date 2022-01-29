//https://github.com/facebook/react/issues/11163#issuecomment-628379291

import { VideoHTMLAttributes, useEffect, useRef } from 'react'
import './Video.css'

type PropsType = VideoHTMLAttributes<HTMLVideoElement> & {
    srcObject: MediaStream | null
}

export default function Video({ srcObject, ...props }: PropsType) {
    const refVideo = useRef<HTMLVideoElement>(null)

    useEffect(() => {
        if (!refVideo.current) return
        refVideo.current.srcObject = srcObject
    }, [srcObject])

    return (
        <div className="videoWrapper">
            <label className="videoTitle">{props.title}</label>
            <video ref={refVideo} {...props} />
        </div>
    )
}