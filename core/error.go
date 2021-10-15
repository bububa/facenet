package core

// ErrCode error code
type ErrCode int

const (
	// UnknownErr unknown error code
	UnknownErr ErrCode = iota
	// ExtractImageSizeTooSmallErr extract image size is too small
	ExtractImageSizeTooSmallErr
	// NoFaceErr no face detected error
	NoFaceErr
	// InferenceFailedErr extract face failed
	InferenceFailedErr
	// ImageToTensorSizeErr convert image to tensor size error
	ImageToTensorSizeErr
	// NegativeDistanceMatchErr match distance is negative
	NegativeDistanceMatchErr
	// TooFarMatchErr match distance is too far
	TooFarMatchErr
	// CollisionMatchErr match distance is larger than collision radius, may need more trainning data
	CollisionMatchErr
	// NothingMatchErr represents nothing matched
	NothingMatchErr
)

// Error custom error object
type Error struct {
	// Code error code
	Code ErrCode `json:"code,omitempty"`
	// Msg error message
	Msg string `json:"msg,omitempty"`
}

// NewError create an error
func NewError(code ErrCode, msg string) Error {
	return Error{
		Code: code,
		Msg:  msg,
	}
}

// Error implement error interface
func (e Error) Error() string {
	return e.Msg
}
