package manifest

import "github.com/asaskevich/govalidator"

func ValidateManifest(manifest any) error {
	if ok, err := govalidator.ValidateStruct(manifest); !ok && err != nil {
		return err
	}

	return nil
}
