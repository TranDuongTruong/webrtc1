

import { useRef, useState } from "react";
import { iceServers } from "../utils/iceServers";

const StreamPublisher = () => {
    const screenVideoRef = useRef(null);
    const cameraVideoRef = useRef(null);
    const [peerConnection, setPeerConnection] = useState(null);
    const [isRecordingAudio, setIsRecordingAudio] = useState(true);

    const startPublishing = async () => {
        try {
            const pc = new RTCPeerConnection(iceServers);
            // 2️⃣ Gửi màn hình với RID là 'screen'
            const screenStream = await navigator.mediaDevices.getDisplayMedia({ video: true });
            const screenTrack = screenStream.getVideoTracks()[0];
            pc.addTransceiver(screenTrack, {
                direction: "sendonly",
                sendEncodings: [{ rid: "screen" }],
            });
            screenVideoRef.current.srcObject = screenStream;
            console.log("✅ Screen track ID:", screenTrack.id);

            // 1️⃣ Gửi camera với RID là 'camera'
            const cameraStream = await navigator.mediaDevices.getUserMedia({ video: true });
            const cameraTrack = cameraStream.getVideoTracks()[0];
            pc.addTransceiver(cameraTrack, {
                direction: "sendonly",
                sendEncodings: [{ rid: "camera" }],
            });
            cameraVideoRef.current.srcObject = cameraStream;
            console.log("✅ Camera track ID:", cameraTrack.id);


            // 3️⃣ Gửi audio (Không cần RID cho audio)
            const audioStream = await navigator.mediaDevices.getUserMedia({ audio: true });
            const audioTrack = audioStream.getAudioTracks()[0];
            pc.addTransceiver(audioTrack, {
                direction: "sendonly",
            });
            console.log("🎤 Added mic audio track");

            const offer = await pc.createOffer();
            await pc.setLocalDescription(offer);

            const response = await fetch("http://localhost:8082/whip", {
                method: "POST",
                body: pc.localDescription.sdp,
                headers: {
                    "Content-Type": "application/sdp",
                    "Accept": "*/*",
                },
                mode: "cors",
            });

            const answer = await response.text();
            await pc.setRemoteDescription({ sdp: answer, type: "answer" });

        } catch (error) {
            console.error("❌ Error in startPublishing:", error);
        }
    };







    return (
        <div>
        <h2>📡 Live Streaming </h2>

            < video ref = { screenVideoRef } autoPlay controls style = {{ width: "100%" }
} />
    < video ref = { cameraVideoRef } autoPlay muted style = {{ width: "150px", height: "100px" }} />

        < button onClick = { startPublishing } >🎥 Start Streaming </button>
            </div>
    );
};

export default StreamPublisher;
