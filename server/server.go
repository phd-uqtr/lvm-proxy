package server

import (
	context "context"
	"flag"
	"fmt"
	"log"
	"net"

	grpc "google.golang.org/grpc"
	api "phd.uqtr.ca/lvm-proxy/api"
	server_pb "phd.uqtr.ca/lvm-proxy/server/pb"
)

var (
	port = flag.Int("port", 50050, "The server port")
)

type server struct {
	server_pb.UnimplementedVolumeServer
	server_pb.UnimplementedVolumeGroupServer
	api *api.LvmProxyApi
}

// ========== Volume group methods ==========
func (s *server) GetVolumeGroupNames(_ context.Context, in *server_pb.GetVolumeGroupNamesRequest) (*server_pb.GetVolumeGroupNamesResponse, error) {
	vg_names := s.api.GetVolumeGroupNames()
	return &server_pb.GetVolumeGroupNamesResponse{
		VolumeGroups: vg_names,
	}, nil
}

func (s *server) GetVolumeGroupInfo(_ context.Context, in *server_pb.GetVolumeGroupInfoRequest) (*server_pb.GetVolumeGroupInfoResponse, error) {
	vgInfo := s.api.GetVolumeGroupInfo(in.GetVolumeGroup())
	return &server_pb.GetVolumeGroupInfoResponse{
		Name:     vgInfo.Name,
		Size:     vgInfo.Size,
		FreeSize: vgInfo.FreeSize,
	}, nil
}

// ========== Volume methods ===========
func (s *server) GetVolumes(_ context.Context, in *server_pb.GetVolumesRequest) (*server_pb.GetVolumesResponse, error) {
	vgName := in.GetVolumeGroup()
	volumes, err := s.api.GetVolumes(vgName)
	if err != nil {
		return &server_pb.GetVolumesResponse{
			Status:  server_pb.Status_FAILURE,
			Message: err.Error(),
		}, nil
	}
	// Map api volumes to response
	var volResponse = make([]*server_pb.CreateLVMVolumeResponse, 0)
	for _, vol := range volumes {
		volResponse = append(volResponse, &server_pb.CreateLVMVolumeResponse{
			VolumeName: vol.Lvo.GetName(),
			VolumePath: vol.DeviceAbsolutionPath,
			BrickPath:  vol.BrickPath,
		})
	}
	return &server_pb.GetVolumesResponse{
		Volumes: volResponse,
		Status:  server_pb.Status_SUCCESS,
	}, nil
}

func (s *server) CreateVolume(_ context.Context, in *server_pb.CreateLVMVolumeRequest) (*server_pb.CreateLVMVolumeResponse, error) {
	volName := in.GetVolumeName()
	vgName := in.GetVolumeGroup()
	size := in.GetSize()
	lvo, err := s.api.CreateVolume(volName, vgName, size)
	if err != nil {
		return &server_pb.CreateLVMVolumeResponse{
			VolumeName: volName,
			VolumePath: "",
			Message:    err.Error(),
			Status:     server_pb.Status_FAILURE,
		}, nil
	}
	return &server_pb.CreateLVMVolumeResponse{
		VolumeName: volName,
		VolumePath: lvo.DeviceAbsolutionPath,
		BrickPath:  lvo.BrickPath,
		Message:    "Volume created successfully",
		Status:     server_pb.Status_SUCCESS,
	}, nil
}

func (s *server) GetVolumeInfo(_ context.Context, in *server_pb.GetVolumeInfoRequest) (*server_pb.GetVolumeInfoResponse, error) {
	volInfo := s.api.GetVolumeInfo(in.GetVolumeGroup(), in.GetVolumeName())
	return &server_pb.GetVolumeInfoResponse{
		VolumeGroup:        volInfo.VolumeGroupName,
		VolumeRelativePath: volInfo.VolumeRelativePath,
		VolumeAbsolutePath: volInfo.VolumeAbsolutePath,
		BrickPath:          volInfo.BrickPath,
		Size:               volInfo.Size,
		FreeSize:           volInfo.FreeSize,
		AllocatedSize:      volInfo.AllocatedSize,
	}, nil
}

func (s *server) DeleteVolume(_ context.Context, in *server_pb.DeleteVolumeRequest) (*server_pb.DeleteVolumeResponse, error) {
	err := s.api.DeleteVolume(in.GetVolumeGroup(), in.GetVolumeName())
	status := server_pb.Status_SUCCESS
	message := "Volume deleted"
	if err != nil {
		status = server_pb.Status_FAILURE
		message = err.Error()
	}
	return &server_pb.DeleteVolumeResponse{
		Status:  status,
		Message: message,
	}, nil
}

func StartServer(api *api.LvmProxyApi) {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	server := &server{
		api: api,
	}
	server_pb.RegisterVolumeServer(s, server)
	server_pb.RegisterVolumeGroupServer(s, server)
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
