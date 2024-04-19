package oidb

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/LagrangeDev/LagrangeGo/packets/pb/service/oidb"
	"github.com/LagrangeDev/LagrangeGo/utils"
	"io"
	"math/rand/v2"
)

func BuildGroupImageUploadReq(groupUin uint32, stream io.Reader) (*OidbPacket, error) {
	// OidbSvcTrpcTcp.0x11c4_100
	data, err := io.ReadAll(stream)
	if err != nil {
		return nil, err
	}
	md5Hash := utils.Md5Digest(data)
	sha1Hash := utils.Sha1Digest(data)
	format, size, err := utils.ImageResolve(data)
	if err != nil {
		return nil, err
	}
	imageExt := utils.GetImageExt(format)

	hexString := "0800180020004a00500062009201009a0100aa010c080012001800200028003a00"
	bytesPbReserveTroop, err := hex.DecodeString(hexString)
	if err != nil {
		return nil, err
	}

	body := &oidb.NTV2RichMediaReq{
		ReqHead: &oidb.MultiMediaReqHead{
			Common: &oidb.CommonHead{
				RequestId: 1,
				Command:   100,
			},
			Scene: &oidb.SceneInfo{
				RequestType:  2,
				BusinessType: 1,
				SceneType:    2,
				Group: &oidb.NTGroupInfo{
					GroupUin: groupUin,
				},
			},
			Client: &oidb.ClientMeta{
				AgentType: 2,
			},
		},
		Upload: &oidb.UploadReq{
			UploadInfo: []*oidb.UploadInfo{
				{
					FileInfo: &oidb.FileInfo{
						FileSize: uint32(len(data)),
						FileHash: fmt.Sprintf("%x", md5Hash),
						FileSha1: fmt.Sprintf("%x", sha1Hash),
						FileName: fmt.Sprintf("%x.%s", md5Hash, imageExt),
						Type: &oidb.FileType{
							Type:        1,
							PicFormat:   uint32(format),
							VideoFormat: 0,
							VoiceFormat: 0,
						},
						Width:    size.X,
						Height:   size.Y,
						Time:     0,
						Original: 1,
					},
					SubFileType: 0,
				},
			},
			TryFastUploadCompleted: true,
			SrvSendMsg:             false,
			ClientRandomId:         rand.Uint64(),
			CompatQMsgSceneType:    2,
			ExtBizInfo: &oidb.ExtBizInfo{
				Pic: &oidb.PicExtBizInfo{
					BytesPbReserveTroop: bytesPbReserveTroop,
				},
				Video: &oidb.VideoExtBizInfo{
					BytesPbReserve: []byte{},
				},
				Ptt: &oidb.PttExtBizInfo{
					BytesReserve:      []byte{},
					BytesPbReserve:    []byte{},
					BytesGeneralFlags: []byte{},
				},
			},
			ClientSeq:       0,
			NoNeedCompatMsg: false,
		},
	}
	return BuildOidbPacket(0x11c4, 100, body, false, true)
}

func ParseGroupImageUploadResp(data []byte) (*oidb.NTV2RichMediaResp, error) { // TODO: return proper response
	var resp oidb.NTV2RichMediaResp
	baseResp, err := ParseOidbPacket(data, &resp)
	if err != nil {
		return nil, err
	}
	if baseResp.ErrorCode != 0 {
		return nil, errors.New(baseResp.ErrorMsg)
	}

	return &resp, nil
}
