package main

import (
	"errors"

	"golang.org/x/net/context"

	"fmt"
	pb "github.com/mpurdon/gomicro-example/user-service/proto/user"
	"golang.org/x/crypto/bcrypt"
)

// Service should implement all of the methods to satisfy the service
// we defined in our protobuf definition. You can check the interface
// in the generated code itself for the exact method signatures etc
// to give you a better idea.
type service struct {
	repo         Repository
	tokenService Authable
}

/*
 * Create a user.
 */
func (s *service) Create(ctx context.Context, req *pb.User, res *pb.Response) error {

	Logger.Infof("Creating user: %v", req)

	// Generates a hashed version of our password
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New(fmt.Sprintf("error hashing password: %v", err))
	}

	req.Password = string(hashedPass)
	if err := s.repo.Create(req); err != nil {
		return errors.New(fmt.Sprintf("error creating user: %v", err))
	}

	token, err := s.tokenService.Encode(req)
	if err != nil {
		return err
	}

	res.User = req
	res.Token = &pb.Token{Token: token}

	/*
	   if err := srv.Publisher.Publish(ctx, req); err != nil {
	       return errors.New(fmt.Sprintf("error publishing event: %v", err))
	   }*/

	return nil
}

/**
 * Get user handler
 */
func (s *service) Get(ctx context.Context, req *pb.User, res *pb.Response) error {

	user, err := s.repo.Get(req)
	if err != nil {
		Logger.Errorf("query error getting user: %v", err)
		return err
	}

	res.User = user

	return nil
}

/**
 * Get all users handler
 */
func (s *service) GetAll(ctx context.Context, req *pb.Request, res *pb.Response) error {
	users, err := s.repo.GetAll()

	if err != nil {
		return err
	}
	res.Users = users

	return nil
}

/**
 * Get a campaign by its GUID
func (repo *CampaignRepository) GetCampaign(guid string) (*pb.Campaign, error) {
    Logger.Infof("Getting campaign with GUID %s from the database.", guid)

    var campaign *pb.Campaign
    campaign.Guid = guid

    if err := repo.db.First(&campaign).Error; err != nil {
        Logger.Errorf("query error getting campaign: %s", err)
        return nil, err
    }

    return campaign, nil
}
*/

/**
 * Handle Authentication
 */
func (s *service) Auth(ctx context.Context, req *pb.User, res *pb.Token) error {
	Logger.Infof("Logging in with: %s|%s", req.Email, req.Password)
	user, err := s.repo.GetByEmail(req.Email)
	if err != nil {
		return err
	}

	// Compares our given password against the hashed password
	// stored in the database
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return err
	}

	token, err := s.tokenService.Encode(user)
	if err != nil {
		return err
	}
	res.Token = token

	return nil
}

/**
 * Validate a given token
 */
func (s *service) ValidateToken(ctx context.Context, req *pb.Token, res *pb.Token) error {

	// Decode token
	claims, err := s.tokenService.Decode(req.Token)

	if err != nil {
		return err
	}

	if claims.User.Id == "" {
		return errors.New("invalid user")
	}

	res.Valid = true

	return nil
}
