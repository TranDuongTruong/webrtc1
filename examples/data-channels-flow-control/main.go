// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

// data-channels-flow-control demonstrates how to use the DataChannel congestion control APIs
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync/atomic"
	"time"

	"github.com/pion/webrtc/v4"
)

const (
	bufferedAmountLowThreshold uint64 = 512 * 1024  // 512 KB
	maxBufferedAmount          uint64 = 1024 * 1024 // 1 MB
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func setRemoteDescription(pc *webrtc.PeerConnection, sdp []byte) {
	var desc webrtc.SessionDescription
	err := json.Unmarshal(sdp, &desc)
	check(err)

	// Apply the desc as the remote description
	err = pc.SetRemoteDescription(desc)
	check(err)
}

func createOfferer() *webrtc.PeerConnection {
	// Prepare the configuration
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{},
	}

	// Create a new PeerConnection
	pc, err := webrtc.NewPeerConnection(config)
	check(err)

	buf := make([]byte, 1024)

	ordered := false
	maxRetransmits := uint16(0)

	options := &webrtc.DataChannelInit{
		Ordered:        &ordered,
		MaxRetransmits: &maxRetransmits,
	}

	sendMoreCh := make(chan struct{}, 1)

	// Create a datachannel with label 'data'
	dataChannel, err := pc.CreateDataChannel("data", options)
	check(err)

	// Register channel opening handling
	dataChannel.OnOpen(func() {
		log.Printf(
			"OnOpen: %s-%d. Start sending a series of 1024-byte packets as fast as it can\n",
			dataChannel.Label(), dataChannel.ID(),
		)

		for {
			err2 := dataChannel.Send(buf)
			check(err2)

			if dataChannel.BufferedAmount() > maxBufferedAmount {
				// Wait until the bufferedAmount becomes lower than the threshold
				<-sendMoreCh
			}
		}
	})

	// Set bufferedAmountLowThreshold so that we can get notified when
	// we can send more
	dataChannel.SetBufferedAmountLowThreshold(bufferedAmountLowThreshold)

	// This callback is made when the current bufferedAmount becomes lower than the threshold
	dataChannel.OnBufferedAmountLow(func() {
		// Make sure to not block this channel or perform long running operations in this callback
		// This callback is executed by pion/sctp. If this callback is blocking it will stop operations
		select {
		case sendMoreCh <- struct{}{}:
		default:
		}
	})

	return pc
}

func createAnswerer() *webrtc.PeerConnection {
	// Prepare the configuration
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{},
	}

	// Create a new PeerConnection
	pc, err := webrtc.NewPeerConnection(config)
	check(err)

	pc.OnDataChannel(func(dataChannel *webrtc.DataChannel) {
		var totalBytesReceived uint64

		// Register channel opening handling
		dataChannel.OnOpen(func() {
			log.Printf("OnOpen: %s-%d. Start receiving data", dataChannel.Label(), dataChannel.ID())
			since := time.Now()

			// Start printing out the observed throughput
			ticker := time.NewTicker(1000 * time.Millisecond)
			defer ticker.Stop()
			for range ticker.C {
				bps := float64(atomic.LoadUint64(&totalBytesReceived)*8) / time.Since(since).Seconds()
				log.Printf("Throughput: %.03f Mbps", bps/1024/1024)
			}
		})

		// Register the OnMessage to handle incoming messages
		dataChannel.OnMessage(func(dcMsg webrtc.DataChannelMessage) {
			n := len(dcMsg.Data)
			atomic.AddUint64(&totalBytesReceived, uint64(n))
		})
	})

	return pc
}

func main() {
	offerPC := createOfferer()
	defer func() {
		if err := offerPC.Close(); err != nil {
			fmt.Printf("cannot close offerPC: %v\n", err)
		}
	}()

	answerPC := createAnswerer()
	defer func() {
		if err := answerPC.Close(); err != nil {
			fmt.Printf("cannot close answerPC: %v\n", err)
		}
	}()

	// Set ICE Candidate handler. As soon as a PeerConnection has gathered a candidate
	// send it to the other peer
	answerPC.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {
			check(offerPC.AddICECandidate(candidate.ToJSON()))
		}
	})

	// Set ICE Candidate handler. As soon as a PeerConnection has gathered a candidate
	// send it to the other peer
	offerPC.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {
			check(answerPC.AddICECandidate(candidate.ToJSON()))
		}
	})

	// Set the handler for Peer connection state
	// This will notify you when the peer has connected/disconnected
	offerPC.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		fmt.Printf("Peer Connection State has changed: %s (offerer)\n", state.String())

		if state == webrtc.PeerConnectionStateFailed {
			// Wait until PeerConnection has had no network activity for 30 seconds or another failure.
			// It may be reconnected using an ICE Restart.
			// Use webrtc.PeerConnectionStateDisconnected if you are interested in detecting faster timeout.
			// Note that the PeerConnection may come back from PeerConnectionStateDisconnected.
			fmt.Println("Peer Connection has gone to failed exiting")
			os.Exit(0)
		}

		if state == webrtc.PeerConnectionStateClosed {
			// PeerConnection was explicitly closed. This usually happens from a DTLS CloseNotify
			fmt.Println("Peer Connection has gone to closed exiting")
			os.Exit(0)
		}
	})

	// Set the handler for Peer connection state
	// This will notify you when the peer has connected/disconnected
	answerPC.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		fmt.Printf("Peer Connection State has changed: %s (answerer)\n", state.String())

		if state == webrtc.PeerConnectionStateFailed {
			// Wait until PeerConnection has had no network activity for 30 seconds or another failure.
			// It may be reconnected using an ICE Restart.
			// Use webrtc.PeerConnectionStateDisconnected if you are interested in detecting faster timeout.
			// Note that the PeerConnection may come back from PeerConnectionStateDisconnected.
			fmt.Println("Peer Connection has gone to failed exiting")
			os.Exit(0)
		}

		if state == webrtc.PeerConnectionStateClosed {
			// PeerConnection was explicitly closed. This usually happens from a DTLS CloseNotify
			fmt.Println("Peer Connection has gone to closed exiting")
			os.Exit(0)
		}
	})

	// Now, create an offer
	offer, err := offerPC.CreateOffer(nil)
	check(err)
	check(offerPC.SetLocalDescription(offer))
	desc, err := json.Marshal(offer)
	check(err)

	setRemoteDescription(answerPC, desc)

	answer, err := answerPC.CreateAnswer(nil)
	check(err)
	check(answerPC.SetLocalDescription(answer))
	desc2, err := json.Marshal(answer)
	check(err)

	setRemoteDescription(offerPC, desc2)

	// Block forever
	select {}
}
