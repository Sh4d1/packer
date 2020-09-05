package scaleway

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

type stepImage struct{}

func (s *stepImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	instanceAPI := instance.NewAPI(state.Get("client").(*scw.Client))
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(*Config)
	snapshotID := state.Get("snapshot_id").(string)
	bootscriptID := ""

	ui.Say(fmt.Sprintf("Creating image: %v", c.ImageName))

	imageResp, err := instanceAPI.GetImage(&instance.GetImageRequest{
		ImageID: c.Image,
	})
	if err != nil {
		err := fmt.Errorf("Error getting initial image info: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if imageResp.Image.DefaultBootscript != nil {
		bootscriptID = imageResp.Image.DefaultBootscript.ID
	}

	createImageResp, err := instanceAPI.CreateImage(&instance.CreateImageRequest{
		Arch:              imageResp.Image.Arch,
		DefaultBootscript: bootscriptID,
		Name:              c.ImageName,
		RootVolume:        snapshotID,
	})
	if err != nil {
		err := fmt.Errorf("Error creating image: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Printf("Image ID: %s", createImageResp.Image.ID)
	state.Put("image_id", createImageResp.Image.ID)
	state.Put("image_name", c.ImageName)
	state.Put("region", c.Zone) // Deprecated
	state.Put("zone", c.Zone)

	return multistep.ActionContinue
}

func (s *stepImage) Cleanup(state multistep.StateBag) {
	// no cleanup
}
