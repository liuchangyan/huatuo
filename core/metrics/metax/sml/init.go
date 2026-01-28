package sml

// sml.Init()
// Init load and initializes the MetaX SML library.
func Init() error {
	if err := libsml.load(); err != nil {
		return err
	}
	return checkReturnCode("mxSmlInit", mxSmlInit())
}

// sml.Shutdown()
// Shutdown shuts down the MetaX SML library.
func Shutdown() error {
	return libsml.close()
}
