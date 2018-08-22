package main

import (
	"golang.org/x/net/context"

	venueProto "github.com/mpurdon/gomicro-example/venue-service/proto/venue"
	//venueProto "../venue-service/proto/venue"
	pb "github.com/mpurdon/gomicro-example/campaign-service/proto/campaign"
)

// Service should implement all of the methods to satisfy the service
// we defined in our protobuf definition. You can check the interface
// in the generated code itself for the exact method signatures etc
// to give you a better idea.
type service struct {
	repo        Repository
	venueClient venueProto.VenueServiceClient
}

/**
 * Create campaigns handler
 */
func (s *service) CreateCampaign(ctx context.Context, req *pb.Campaign, res *pb.Response) error {

	Logger.Infof("Attempting to create campaign %s\n", req.Name)

	var capacity int32
	if len(req.Rewards) > 0 {
		capacity = int32(req.Rewards[0].Available)
	}

	response, err := s.venueClient.FindAvailable(context.Background(), &venueProto.VenueSpecification{
		Location: req.Location,
		Capacity: capacity,
	})

	if err != nil {
		Logger.Warnf("Could not find a venue for the campaign: %s", err)
		return err
	}

	Logger.Infof("Found venue: %s, setting campaign venue to %s \n", response.Venue.Name, response.Venue.Id)

	// We set the VesselId as the vessel we got back from our
	// vessel service
	req.VenueId = response.Venue.Id

	// Save our campaign
	campaign, err := s.repo.Create(req)
	if err != nil {
		Logger.Errorf("Could not create campaign: %s", err)
		return err
	}

	res.Created = true
	res.Campaign = campaign

	return nil
}

/**
 * Get campaigns handler
 */
func (s *service) GetCampaigns(ctx context.Context, req *pb.GetRequest, res *pb.Response) error {
	campaigns, err := s.repo.GetAll()

	if err != nil {
		return err
	}
	res.Campaigns = campaigns

	return nil
}
