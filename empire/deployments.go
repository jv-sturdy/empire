package empire

import (
	"fmt"

	"golang.org/x/net/context"
)

// Deployment statuses.
const (
	StatusPending = "pending"
	StatusFailed  = "failed"
	StatusSuccess = "success"
)

// DeploymentsCreateOpts represents options that can be passed when creating a
// new Deployment.
type DeploymentsCreateOpts struct {
	// App is the app that is being deployed to.
	App *App

	// Image is the image that's being deployed.
	Image Image

	// EventCh will receive deployment events during deployment.
	EventCh chan Event
}

type deployer struct {
	*appsService
	*configsService
	*slugsService
	*releasesService
}

// DeploymentsDo performs the Deployment.
func (s *deployer) DeploymentsDo(ctx context.Context, opts DeploymentsCreateOpts) (*Release, error) {
	app, image := opts.App, opts.Image

	// Grab the latest config.
	config, err := s.ConfigsCurrent(app)
	if err != nil {
		return nil, err
	}

	// Create a new slug for the docker image.
	slug, err := s.SlugsCreateByImage(image, opts.EventCh)
	if err != nil {
		return nil, err
	}

	// Create a new release for the Config
	// and Slug.
	desc := fmt.Sprintf("Deploy %s", image.String())
	return s.ReleasesCreate(ctx, &Release{
		App:         app,
		Config:      config,
		Slug:        slug,
		Description: desc,
	})
}

func (s *deployer) DeployImageToApp(ctx context.Context, app *App, image Image, out chan Event) (*Release, error) {
	if err := s.appsService.AppsEnsureRepo(app, image.Repo); err != nil {
		return nil, err
	}

	return s.DeploymentsDo(ctx, DeploymentsCreateOpts{
		App:     app,
		Image:   image,
		EventCh: out,
	})
}

// Deploy deploys an Image to the cluster.
func (s *deployer) DeployImage(ctx context.Context, image Image, out chan Event) (*Release, error) {
	app, err := s.appsService.AppsFindOrCreateByRepo(image.Repo)
	if err != nil {
		return nil, err
	}

	return s.DeployImageToApp(ctx, app, image, out)
}
