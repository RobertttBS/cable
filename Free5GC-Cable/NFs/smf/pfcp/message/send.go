package message

import (
	"net"
	//	"fmt"

	"github.com/free5gc/pfcp"
	"github.com/free5gc/pfcp/pfcpType"
	"github.com/free5gc/pfcp/pfcpUdp"
	"github.com/free5gc/smf/context"
	"github.com/free5gc/smf/logger"
	"github.com/free5gc/smf/pfcp/udp"

	"github.com/fatih/structs"
	"github.com/free5gc/smf/monitorhandler/mongolib"
	"go.mongodb.org/mongo-driver/bson"
)

var seq uint32

func getSeqNumber() uint32 {
	seq++
	return seq
}

func SendPfcpAssociationSetupRequest(upNodeID pfcpType.NodeID) {
	pfcpMsg, err := BuildPfcpAssociationSetupRequest()
	if err != nil {
		logger.PfcpLog.Errorf("Build PFCP Association Setup Request failed: %v", err)
		return
	}

	message := pfcp.Message{
		Header: pfcp.Header{
			Version:        pfcp.PfcpVersion,
			MP:             0,
			S:              pfcp.SEID_NOT_PRESENT,
			MessageType:    pfcp.PFCP_ASSOCIATION_SETUP_REQUEST,
			SequenceNumber: getSeqNumber(),
		},
		Body: pfcpMsg,
	}

	addr := &net.UDPAddr{
		IP:   upNodeID.ResolveNodeIdToIp(),
		Port: pfcpUdp.PFCP_PORT,
	}

	udp.SendPfcp(message, addr)
}

func SendPfcpAssociationSetupResponse(upNodeID pfcpType.NodeID, cause pfcpType.Cause) {
	pfcpMsg, err := BuildPfcpAssociationSetupResponse(cause)
	if err != nil {
		logger.PfcpLog.Errorf("Build PFCP Association Setup Response failed: %v", err)
		return
	}

	message := pfcp.Message{
		Header: pfcp.Header{
			Version:        pfcp.PfcpVersion,
			MP:             0,
			S:              pfcp.SEID_NOT_PRESENT,
			MessageType:    pfcp.PFCP_ASSOCIATION_SETUP_RESPONSE,
			SequenceNumber: 1,
		},
		Body: pfcpMsg,
	}

	addr := &net.UDPAddr{
		IP:   upNodeID.ResolveNodeIdToIp(),
		Port: pfcpUdp.PFCP_PORT,
	}

	udp.SendPfcp(message, addr)
}

func SendPfcpAssociationReleaseRequest(upNodeID pfcpType.NodeID) {
	pfcpMsg, err := BuildPfcpAssociationReleaseRequest()
	if err != nil {
		logger.PfcpLog.Errorf("Build PFCP Association Release Request failed: %v", err)
		return
	}

	message := pfcp.Message{
		Header: pfcp.Header{
			Version:        pfcp.PfcpVersion,
			MP:             0,
			S:              pfcp.SEID_NOT_PRESENT,
			MessageType:    pfcp.PFCP_ASSOCIATION_RELEASE_REQUEST,
			SequenceNumber: 1,
		},
		Body: pfcpMsg,
	}

	addr := &net.UDPAddr{
		IP:   upNodeID.ResolveNodeIdToIp(),
		Port: pfcpUdp.PFCP_PORT,
	}

	udp.SendPfcp(message, addr)
}

func SendPfcpAssociationReleaseResponse(upNodeID pfcpType.NodeID, cause pfcpType.Cause) {
	pfcpMsg, err := BuildPfcpAssociationReleaseResponse(cause)
	if err != nil {
		logger.PfcpLog.Errorf("Build PFCP Association Release Response failed: %v", err)
		return
	}

	message := pfcp.Message{
		Header: pfcp.Header{
			Version:        pfcp.PfcpVersion,
			MP:             0,
			S:              pfcp.SEID_NOT_PRESENT,
			MessageType:    pfcp.PFCP_ASSOCIATION_RELEASE_RESPONSE,
			SequenceNumber: 1,
		},
		Body: pfcpMsg,
	}

	addr := &net.UDPAddr{
		IP:   upNodeID.ResolveNodeIdToIp(),
		Port: pfcpUdp.PFCP_PORT,
	}

	udp.SendPfcp(message, addr)
}

type PDUSessionInfo struct {
	UPFip          string `json:"upfip" structs:"upfip"`
	SequenceNumber uint32 `json:"seqnum" structs:"seqnum"`
	TEID           uint32 `json:"teid" structs:"teid"`
	UEIP           string `json:"ueip" structs:"ueip"`
	PDUSessionID   int    `json:"pdusessionid" structs:"pdusessionid"`
}

func SendPfcpSessionEstablishmentRequest(
	upNodeID pfcpType.NodeID,
	ctx *context.SMContext,
	pdrList []*context.PDR, farList []*context.FAR, barList []*context.BAR, qerList []*context.QER) {
	pfcpMsg, err := BuildPfcpSessionEstablishmentRequest(upNodeID, ctx, pdrList, farList, barList, qerList)
	if err != nil {
		logger.PfcpLog.Errorf("Build PFCP Session Establishment Request failed: %v", err)
		return
	}

	message := pfcp.Message{
		Header: pfcp.Header{
			Version:         pfcp.PfcpVersion,
			MP:              1,
			S:               pfcp.SEID_PRESENT,
			MessageType:     pfcp.PFCP_SESSION_ESTABLISHMENT_REQUEST,
			SEID:            0,
			SequenceNumber:  getSeqNumber(),
			MessagePriority: 0,
		},
		Body: pfcpMsg,
	}

	upfip := upNodeID.ResolveNodeIdToIp().String()
	seqnum := message.Header.SequenceNumber
	teid := pfcpMsg.CreatePDR[0].PDI.LocalFTEID.Teid
	ueip := pfcpMsg.CreatePDR[0].PDI.UEIPAddress.Ipv4Address
	pdusessionid := 0

	//	fmt.Printf("%v, %v, %v\n", seqnum, teid, ueip)
	//	fmt.Printf("%s\n", ueip.String())
	perfclient := mongolib.PerfClient{"mongodb://192.168.56.155:27017", "PDUSessionInfo", "PDUs"}
	perfclient.ConnectDB(perfclient.DBurl, perfclient.DBname)

	postData := PDUSessionInfo{upfip, seqnum, teid, ueip.String(), pdusessionid}
	postDataMap := structs.Map(&postData)
	filter := bson.M{"seqnum": seqnum}
	getresult := perfclient.PostOne(perfclient.Collname, filter, postDataMap)
	if !getresult {
		logger.PfcpLog.Warnf("Post a New PDU Session Info to MongoDB succeed!")
	} else {
		logger.PfcpLog.Infof("Update a PDU Session Info succeed!")
	}

	upaddr := &net.UDPAddr{
		IP:   upNodeID.ResolveNodeIdToIp(),
		Port: pfcpUdp.PFCP_PORT,
	}
	logger.PduSessLog.Traceln("[SMF] Send SendPfcpSessionEstablishmentRequest")
	logger.PduSessLog.Traceln("Send to addr ", upaddr.String())

	udp.SendPfcp(message, upaddr)
}

// Deprecated: PFCP Session Establishment Procedure should be initiated by the CP function
func SendPfcpSessionEstablishmentResponse(addr *net.UDPAddr) {
	pfcpMsg, err := BuildPfcpSessionEstablishmentResponse()
	if err != nil {
		logger.PfcpLog.Errorf("Build PFCP Session Establishment Response failed: %v", err)
		return
	}

	message := pfcp.Message{
		Header: pfcp.Header{
			Version:         pfcp.PfcpVersion,
			MP:              1,
			S:               pfcp.SEID_PRESENT,
			MessageType:     pfcp.PFCP_SESSION_ESTABLISHMENT_RESPONSE,
			SEID:            123456789123456789,
			SequenceNumber:  1,
			MessagePriority: 12,
		},
		Body: pfcpMsg,
	}

	udp.SendPfcp(message, addr)
}

func SendPfcpSessionModificationRequest(upNodeID pfcpType.NodeID,
	ctx *context.SMContext,
	pdrList []*context.PDR, farList []*context.FAR, barList []*context.BAR, qerList []*context.QER) (seqNum uint32) {
	pfcpMsg, err := BuildPfcpSessionModificationRequest(upNodeID, ctx, pdrList, farList, barList, qerList)
	if err != nil {
		logger.PfcpLog.Errorf("Build PFCP Session Modification Request failed: %v", err)
		return
	}

	seqNum = getSeqNumber()
	nodeIDtoIP := upNodeID.ResolveNodeIdToIp().String()
	remoteSEID := ctx.PFCPContext[nodeIDtoIP].RemoteSEID
	message := pfcp.Message{
		Header: pfcp.Header{
			Version:         pfcp.PfcpVersion,
			MP:              1,
			S:               pfcp.SEID_PRESENT,
			MessageType:     pfcp.PFCP_SESSION_MODIFICATION_REQUEST,
			SEID:            remoteSEID,
			SequenceNumber:  seqNum,
			MessagePriority: 12,
		},
		Body: pfcpMsg,
	}

	upaddr := &net.UDPAddr{
		IP:   upNodeID.ResolveNodeIdToIp(),
		Port: pfcpUdp.PFCP_PORT,
	}

	udp.SendPfcp(message, upaddr)
	return seqNum
}

// Deprecated: PFCP Session Modification Procedure should be initiated by the CP function
func SendPfcpSessionModificationResponse(addr *net.UDPAddr) {
	pfcpMsg, err := BuildPfcpSessionModificationResponse()
	if err != nil {
		logger.PfcpLog.Errorf("Build PFCP Session Modification Response failed: %v", err)
		return
	}

	message := pfcp.Message{
		Header: pfcp.Header{
			Version:         pfcp.PfcpVersion,
			MP:              1,
			S:               pfcp.SEID_PRESENT,
			MessageType:     pfcp.PFCP_SESSION_MODIFICATION_RESPONSE,
			SEID:            123456789123456789,
			SequenceNumber:  1,
			MessagePriority: 12,
		},
		Body: pfcpMsg,
	}

	udp.SendPfcp(message, addr)
}

func SendPfcpSessionDeletionRequest(upNodeID pfcpType.NodeID, ctx *context.SMContext) (seqNum uint32) {
	pfcpMsg, err := BuildPfcpSessionDeletionRequest()
	if err != nil {
		logger.PfcpLog.Errorf("Build PFCP Session Deletion Request failed: %v", err)
		return
	}
	seqNum = getSeqNumber()
	nodeIDtoIP := upNodeID.ResolveNodeIdToIp().String()
	remoteSEID := ctx.PFCPContext[nodeIDtoIP].RemoteSEID
	message := pfcp.Message{
		Header: pfcp.Header{
			Version:         pfcp.PfcpVersion,
			MP:              1,
			S:               pfcp.SEID_PRESENT,
			MessageType:     pfcp.PFCP_SESSION_DELETION_REQUEST,
			SEID:            remoteSEID,
			SequenceNumber:  seqNum,
			MessagePriority: 12,
		},
		Body: pfcpMsg,
	}

	upaddr := &net.UDPAddr{
		IP:   upNodeID.ResolveNodeIdToIp(),
		Port: pfcpUdp.PFCP_PORT,
	}

	udp.SendPfcp(message, upaddr)

	return seqNum
}

// Deprecated: PFCP Session Deletion Procedure should be initiated by the CP function
func SendPfcpSessionDeletionResponse(addr *net.UDPAddr) {
	pfcpMsg, err := BuildPfcpSessionDeletionResponse()
	if err != nil {
		logger.PfcpLog.Errorf("Build PFCP Session Deletion Response failed: %v", err)
		return
	}

	message := pfcp.Message{
		Header: pfcp.Header{
			Version:         pfcp.PfcpVersion,
			MP:              1,
			S:               pfcp.SEID_PRESENT,
			MessageType:     pfcp.PFCP_SESSION_DELETION_RESPONSE,
			SEID:            123456789123456789,
			SequenceNumber:  1,
			MessagePriority: 12,
		},
		Body: pfcpMsg,
	}

	udp.SendPfcp(message, addr)
}

func SendPfcpSessionReportResponse(addr *net.UDPAddr, cause pfcpType.Cause, seqFromUPF uint32, SEID uint64) {
	pfcpMsg, err := BuildPfcpSessionReportResponse(cause)
	if err != nil {
		logger.PfcpLog.Errorf("Build PFCP Session Report Response failed: %v", err)
		return
	}

	message := pfcp.Message{
		Header: pfcp.Header{
			Version:        pfcp.PfcpVersion,
			MP:             0,
			S:              pfcp.SEID_PRESENT,
			MessageType:    pfcp.PFCP_SESSION_REPORT_RESPONSE,
			SequenceNumber: seqFromUPF,
			SEID:           SEID,
		},
		Body: pfcpMsg,
	}

	udp.SendPfcp(message, addr)
}

func SendHeartbeatResponse(addr *net.UDPAddr, seq uint32) {
	pfcpMsg := pfcp.HeartbeatResponse{
		RecoveryTimeStamp: &pfcpType.RecoveryTimeStamp{
			RecoveryTimeStamp: udp.ServerStartTime,
		},
	}

	message := pfcp.Message{
		Header: pfcp.Header{
			Version:        pfcp.PfcpVersion,
			MP:             0,
			S:              pfcp.SEID_NOT_PRESENT,
			MessageType:    pfcp.PFCP_HEARTBEAT_RESPONSE,
			SequenceNumber: seq,
		},
		Body: pfcpMsg,
	}

	udp.SendPfcp(message, addr)
}

func SendPfcpUpfReportRequest(upNodeID pfcpType.NodeID) {
	pfcpMsg, err := BuildPfcpUpfReportRequest()
	if err != nil {
		logger.PfcpLog.Errorf("Build PFCP UPF Report Request failed: %v", err)
		return
	}

	message := pfcp.Message{
		Header: pfcp.Header{
			Version:        pfcp.PfcpVersion,
			MP:             0,
			S:              pfcp.SEID_NOT_PRESENT,
			MessageType:    pfcp.PFCP_UPF_REPORT_REQUEST,
			SequenceNumber: getSeqNumber(),
		},
		Body: pfcpMsg,
	}

	addr := &net.UDPAddr{
		IP:   upNodeID.ResolveNodeIdToIp(),
		Port: pfcpUdp.PFCP_PORT,
	}

	udp.SendPfcp(message, addr)
}

func SendPduSessionDataReportRequest(upNodeID pfcpType.NodeID) {
	pfcpMsg, err := BuildPduSessionDataReportRequest()
	if err != nil {
		logger.PfcpLog.Errorf("Build PDU Session Data Report Request failed: %v", err)
		return
	}

	message := pfcp.Message{
		Header: pfcp.Header{
			Version:        pfcp.PfcpVersion,
			MP:             0,
			S:              pfcp.SEID_NOT_PRESENT,
			MessageType:    pfcp.PDU_SESSION_DATA_REPORT_REQUEST,
			SequenceNumber: getSeqNumber(),
		},
		Body: pfcpMsg,
	}

	addr := &net.UDPAddr{
		IP:   upNodeID.ResolveNodeIdToIp(),
		Port: pfcpUdp.PFCP_PORT,
	}

	udp.SendPfcp(message, addr)
}

func SendPDUSessionReleaseFromSMFRequest(smfNodeID pfcpType.NodeID, SEID uint64) {
	// this function is not being called
	pduSessionID := pfcpType.PDUSessionID{
		PduSessionid: smfNodeID.NodeIdValue,
	}
	pfcpMsg, err := BuildPDUSessionReleaseFromSMFRequest(pduSessionID) // PDUSessionID is []byte, use smfNodeID.NodeIdValue to replace
	if err != nil {
		logger.PfcpLog.Errorf("Build PDU Session Release From SMF Request failed: %v", err)
		return
	}

	message := pfcp.Message{
		Header: pfcp.Header{
			Version:        pfcp.PfcpVersion,
			MP:             0,
			S:              pfcp.SEID_PRESENT,
			MessageType:    pfcp.PDU_SESSION_RELEASE_FROM_SMF_REQUEST,
			SequenceNumber: getSeqNumber(),
			SEID:           SEID,
		},
		Body: pfcpMsg,
	}

	addr := &net.UDPAddr{
		IP:   smfNodeID.ResolveNodeIdToIp(),
		Port: pfcpUdp.PFCP_PORT,
	}

	udp.SendPfcp(message, addr)
}

func SendPDUSessionReleaseFromSMFResponse(smfNodeID pfcpType.NodeID, cause pfcpType.Cause, seq uint32, SEID uint64) {
	pfcpMsg, err := BuildPDUSessionReleaseFromSMFResponse(cause)
	if err != nil {
		logger.PfcpLog.Errorf("Build PDU Session Release From SMF Response failed: %v", err)
		return
	}
	message := pfcp.Message{
		Header: pfcp.Header{
			Version:        pfcp.PfcpVersion,
			MP:             0,
			S:              pfcp.SEID_PRESENT,
			MessageType:    pfcp.PDU_SESSION_RELEASE_FROM_SMF_RESPONSE,
			SequenceNumber: seq,
			SEID:           SEID,
		},
		Body: pfcpMsg,
	}

	addr := &net.UDPAddr{
		IP:   smfNodeID.ResolveNodeIdToIp(),
		Port: pfcpUdp.PFCP_PORT,
	}

	udp.SendPfcp(message, addr)
}
