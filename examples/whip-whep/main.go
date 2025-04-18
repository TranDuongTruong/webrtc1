

// package main
// import (
//     "fmt"
//     "io"
//     "net/http"
//     "time"

//     "github.com/pion/interceptor"
//     "github.com/pion/webrtc/v4"
//     rtcp "github.com/pion/rtcp" // Dùng alias rtcp
// )
// func enableCORS(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Access-Control-Allow-Origin", "*")
// 	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
// 	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
// 	w.Header().Set("Access-Control-Expose-Headers", "Location")
// 	if r != nil && r.Method == "OPTIONS" {
// 		w.WriteHeader(http.StatusOK)
// 		return
// 	}
// }

// var (
// 	cameraTrack *webrtc.TrackLocalStaticRTP
// 	audioTrack  *webrtc.TrackLocalStaticRTP
// 	peerConnectionConfiguration = webrtc.Configuration{
// 		ICEServers: []webrtc.ICEServer{
// 			{
// 				URLs: []string{"stun:stun.l.google.com:19302", "stun:stun1.l.google.com:19302"},
// 			},
// 			{
// 				URLs:       []string{"turn:relay1.expressturn.com:3478"},
// 				Username:   "ef5d6f0c",
// 				Credential: "nf4A2v6p",
// 			},
// 		},
// 	}
// )

// func main() {
// 	var err error
// 	cameraTrack, err = webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{
// 		MimeType: webrtc.MimeTypeVP8,
// 	}, "screen", "pion")
// 	if err != nil {
// 		panic(err)
// 	}

// 	audioTrack, err = webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{
// 		MimeType: webrtc.MimeTypeOpus,
// 	}, "audio", "pion")
// 	if err != nil {
// 		panic(err)
// 	}

// 	http.HandleFunc("/whip", whipHandler)
// 	http.HandleFunc("/whep", whepHandler)

// 	fmt.Println("Server running on http://localhost:8083")
// 	http.ListenAndServe(":8083", nil)
// }

// func whipHandler(res http.ResponseWriter, req *http.Request) {
// 	enableCORS(res, req)

// 	offer, err := io.ReadAll(req.Body)
// 	if err != nil {
// 		http.Error(res, "Failed to read offer", http.StatusBadRequest)
// 		return
// 	}

// 	mediaEngine := &webrtc.MediaEngine{}
// 	mediaEngine.RegisterDefaultCodecs()

// 	interceptorRegistry := &interceptor.Registry{}
// 	api := webrtc.NewAPI(webrtc.WithMediaEngine(mediaEngine), webrtc.WithInterceptorRegistry(interceptorRegistry))

// 	peerConnection, err := api.NewPeerConnection(peerConnectionConfiguration)
// 	if err != nil {
// 		http.Error(res, "Failed to create PeerConnection", http.StatusInternalServerError)
// 		return
// 	}

	

// 	peerConnection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
// 		go sendPLI(receiver, peerConnection)
	
// 		rtpBuf := make([]byte, 1400)
// 		var localTrack *webrtc.TrackLocalStaticRTP
	
// 		// Xác định loại track là video hay audio
// 		if track.Kind() == webrtc.RTPCodecTypeVideo {
// 			localTrack = cameraTrack
// 			fmt.Println("✅ Assigned to screen track\t", track.ID())
// 		} else if track.Kind() == webrtc.RTPCodecTypeAudio {
// 			localTrack = audioTrack
// 			fmt.Println("✅ Assigned to audio track", track.ID())
// 		}
	
// 		// Nếu track không xác định, bỏ qua
// 		if localTrack == nil {
// 			fmt.Println("⚠️ Unknown track type")
// 			return
// 		}
	
// 		// Đọc dữ liệu từ track và ghi vào local track
// 		for {
// 			n, _, readErr := track.Read(rtpBuf)
// 			if readErr != nil {
// 				fmt.Println("❌ Error reading from track:", readErr)
// 				break
// 			}
// 			if _, writeErr := localTrack.Write(rtpBuf[:n]); writeErr != nil {
// 				fmt.Println("❌ Error writing to local track:", writeErr)
// 				break
// 			}
// 		}
// 	})
	
// 	writeAnswer(res, peerConnection, offer, "/whip")
// }

// func sendPLI(receiver *webrtc.RTPReceiver, pc *webrtc.PeerConnection) {
// 	ticker := time.NewTicker(500 * time.Millisecond) // Gửi PLI mỗi 500ms
// 	defer ticker.Stop()

// 	for range ticker.C {
// 		err := pc.WriteRTCP([]rtcp.Packet{
// 			&rtcp.PictureLossIndication{
// 				MediaSSRC: uint32(receiver.Track().SSRC()),
// 			},
// 		})
// 				if err != nil {
// 			fmt.Println("❌ Error sending PLI:", err)
// 			return
// 		}
// 	}
// }

// func whepHandler(res http.ResponseWriter, req *http.Request) {
// 	enableCORS(res, req)

// 	offer, err := io.ReadAll(req.Body)
// 	if err != nil {
// 		http.Error(res, "Failed to read offer", http.StatusBadRequest)
// 		return
// 	}

// 	peerConnection, err := webrtc.NewPeerConnection(peerConnectionConfiguration)
// 	if err != nil {
// 		http.Error(res, "Failed to create PeerConnection", http.StatusInternalServerError)
// 		return
// 	}

// 	senders := []*webrtc.RTPSender{}
// 	for _, track := range []*webrtc.TrackLocalStaticRTP{cameraTrack, audioTrack} {
// 		sender, err := peerConnection.AddTrack(track)
// 		if err != nil {
// 			fmt.Printf("❌ Failed to add track: %v\n", err)
// 			continue
// 		}
// 		senders = append(senders, sender)
// 	}

// 	writeAnswer(res, peerConnection, offer, "/whep")
// }

// func writeAnswer(res http.ResponseWriter, peerConnection *webrtc.PeerConnection, offer []byte, path string) {
// 	if err := peerConnection.SetRemoteDescription(webrtc.SessionDescription{
// 		Type: webrtc.SDPTypeOffer,
// 		SDP:  string(offer),
// 	}); err != nil {
// 		http.Error(res, "Failed to set remote description", http.StatusInternalServerError)
// 		return
// 	}

// 	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
// 	answer, _ := peerConnection.CreateAnswer(nil)
// 	peerConnection.SetLocalDescription(answer)
// 	<-gatherComplete

// 	res.Header().Add("Location", path)
// 	res.WriteHeader(http.StatusCreated)
// 	fmt.Fprint(res, peerConnection.LocalDescription().SDP)
// }






package main

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pion/interceptor"
	"github.com/pion/webrtc/v4"
	rtcp "github.com/pion/rtcp" // Alias for rtcp package
)

// Cấu hình CORS (Cross-Origin Resource Sharing) cho phép các yêu cầu từ các nguồn khác
func enableCORS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Expose-Headers", "Location")
	if r != nil && r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
}

// Biến toàn cục để lưu track của mỗi stream theo streamId
var streamTracks = make(map[string]map[string]*webrtc.TrackLocalStaticRTP)

// Cấu hình PeerConnection
var peerConnectionConfiguration = webrtc.Configuration{
	ICEServers: []webrtc.ICEServer{
		{
			URLs: []string{"stun:stun.l.google.com:19302", "stun:stun1.l.google.com:19302"},
		},
		{
			URLs:       []string{"turn:relay1.expressturn.com:3478"},
			Username:   "ef5d6f0c",
			Credential: "nf4A2v6p",
		},
	},
}

// Hàm chính để khởi tạo các track và cấu hình server
func main() {
	http.HandleFunc("/whip", whipHandler)  // Publisher gọi vào /whip
	http.HandleFunc("/whep", whepHandler)  // Viewer gọi vào /whep

	fmt.Println("Server running on http://localhost:8084")
	http.ListenAndServe(":8084", nil)
}

// Xử lý yêu cầu từ Publisher khi gọi vào /whip
func whipHandler(res http.ResponseWriter, req *http.Request) {
	enableCORS(res, req)

	// Đọc offer từ request body
	offer, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "Failed to read offer", http.StatusBadRequest)
		return
	}

	// Tạo PeerConnection
	mediaEngine := &webrtc.MediaEngine{}
	mediaEngine.RegisterDefaultCodecs()
	interceptorRegistry := &interceptor.Registry{}
	api := webrtc.NewAPI(webrtc.WithMediaEngine(mediaEngine), webrtc.WithInterceptorRegistry(interceptorRegistry))

	peerConnection, err := api.NewPeerConnection(peerConnectionConfiguration)
	if err != nil {
		http.Error(res, "Failed to create PeerConnection", http.StatusInternalServerError)
		return
	}

	// Lưu track khi nhận được từ publisher
	peerConnection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		go sendPLI(receiver, peerConnection)

		rtpBuf := make([]byte, 1400)
		var localTrack *webrtc.TrackLocalStaticRTP

		// Xác định loại track (video hoặc audio) và lưu vào map
		if track.Kind() == webrtc.RTPCodecTypeVideo {
			localTrack = createVideoTrack() // Tạo track video
			fmt.Println("✅ Assigned to screen track", track.ID())
		} else if track.Kind() == webrtc.RTPCodecTypeAudio {
			localTrack = createAudioTrack() // Tạo track audio
			fmt.Println("✅ Assigned to audio track", track.ID())
		}

		// Lưu track vào map với streamId
		streamId := req.URL.Query().Get("streamId")
		saveStreamTracks(streamId, localTrack, track)

		// Đọc dữ liệu từ track và ghi vào local track
		for {
			n, _, readErr := track.Read(rtpBuf)
			if readErr != nil {
				fmt.Println("❌ Error reading from track:", readErr)
				break
			}
			if _, writeErr := localTrack.Write(rtpBuf[:n]); writeErr != nil {
				fmt.Println("❌ Error writing to local track:", writeErr)
				break
			}
		}
	})

	// Trả về SDP cho publisher
	writeAnswer(res, peerConnection, offer, "/whip")
}

// Lưu track vào map theo streamId
func saveStreamTracks(streamId string, localTrack *webrtc.TrackLocalStaticRTP, track *webrtc.TrackRemote) {
	if _, exists := streamTracks[streamId]; !exists {
		streamTracks[streamId] = make(map[string]*webrtc.TrackLocalStaticRTP)
	}
	if track.Kind() == webrtc.RTPCodecTypeVideo {
		streamTracks[streamId]["video"] = localTrack
	} else if track.Kind() == webrtc.RTPCodecTypeAudio {
		streamTracks[streamId]["audio"] = localTrack
	}
}

// Tạo track video mới
func createVideoTrack() *webrtc.TrackLocalStaticRTP {
	// Tạo track video mới
	videoTrack, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{
		MimeType: webrtc.MimeTypeVP8,
	}, "video", "pion")
	if err != nil {
		panic(err)
	}
	return videoTrack
}

// Tạo track audio mới
func createAudioTrack() *webrtc.TrackLocalStaticRTP {
	// Tạo track audio mới
	audioTrack, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{
		MimeType: webrtc.MimeTypeOpus,
	}, "audio", "pion")
	if err != nil {
		panic(err)
	}
	return audioTrack
}

// Xử lý yêu cầu từ Viewer khi gọi vào /whep
func whepHandler(res http.ResponseWriter, req *http.Request) {
	enableCORS(res, req)

	// Lấy streamId từ query parameter
	streamId := req.URL.Query().Get("streamId")

	// Lấy track cho streamId
	tracks, err := getStreamTracks(streamId)
	if err != nil {
		http.Error(res, err.Error(), http.StatusNotFound)
		return
	}

	// Tạo PeerConnection cho viewer
	mediaEngine := &webrtc.MediaEngine{}
	mediaEngine.RegisterDefaultCodecs()
	// interceptorRegistry := &interceptor.Registry{}
	// api := webrtc.NewAPI(webrtc.WithMediaEngine(mediaEngine), webrtc.WithInterceptorRegistry(interceptorRegistry))

	peerConnection, err := webrtc.NewPeerConnection(peerConnectionConfiguration)
	if err != nil {
		http.Error(res, "Failed to create PeerConnection", http.StatusInternalServerError)
		return
	}

	// Đăng ký track video và audio cho PeerConnection
	for trackKind, track := range tracks {
		// Kiểm tra và đăng ký track tương ứng (video/audio)
		_, err := peerConnection.AddTrack(track)
		if err != nil {
			fmt.Println("❌ Failed to add", trackKind, "track:", err)
			continue
		}
	}

	// Trả về SDP cho viewer
	offer, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "Failed to read offer", http.StatusBadRequest)
		return
	}
	writeAnswer(res, peerConnection, offer, "/whep")
}

// Lấy track từ streamTracks khi viewer yêu cầu
func getStreamTracks(streamId string) (map[string]*webrtc.TrackLocalStaticRTP, error) {
	// Kiểm tra xem streamId có tồn tại không
	tracks, exists := streamTracks[streamId]
	if !exists {
		return nil, fmt.Errorf("Stream not found: %s", streamId)
	}
	return tracks, nil
}

// Trả về SDP cho viewer hoặc publisher
func writeAnswer(res http.ResponseWriter, peerConnection *webrtc.PeerConnection, offer []byte, path string) {
	if err := peerConnection.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  string(offer),
	}); err != nil {
		http.Error(res, "Failed to set remote description", http.StatusInternalServerError)
		return
	}

	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
	answer, _ := peerConnection.CreateAnswer(nil)
	peerConnection.SetLocalDescription(answer)
	<-gatherComplete

	res.Header().Add("Location", path)
	res.WriteHeader(http.StatusCreated)
	fmt.Fprint(res, peerConnection.LocalDescription().SDP)
}

// Gửi Picture Loss Indication (PLI) để yêu cầu phát lại hình ảnh khi bị mất gói
func sendPLI(receiver *webrtc.RTPReceiver, pc *webrtc.PeerConnection) {
	ticker := time.NewTicker(2000 * time.Millisecond) // Gửi PLI mỗi 2 giây
	defer ticker.Stop()

	for range ticker.C {
		err := pc.WriteRTCP([]rtcp.Packet{
			&rtcp.PictureLossIndication{
				MediaSSRC: uint32(receiver.Track().SSRC()),
			},
		})
		if err != nil {
			fmt.Println("❌ Error sending PLI:", err)
			return
		}
	}
}
