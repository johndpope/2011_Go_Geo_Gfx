package pipeline

import (
	"log"
	"math"
	"time"

	num "tshared/numutil"
)

var (
	CamPos, CamLook, CamRot, CamRad, CamSin, CamCos num.Vec3
	CamTurnLeft, CamTurnRight, CamTurnUp, CamTurnDown, CamMoveBack, CamMoveFwd, CamMoveLeft, CamMoveRight, CamMoveUp, CamMoveDown float64
	CamMove, CamTurn bool = false, false
	CamMoveSpeedMps float64 = 8
	CamTurnSpeedDps float64 = 180
	TimeLast, TimeNow time.Time
	TimeSecsElapsed float64
)

var (
	camMoveSpeed, camTurnSpeed float64
)

const (
	KB = 1024
	MB = KB * KB
	GB = MB * KB
)

func CleanUp (forReinit bool) {
}

func PreRender () {
	if CamTurn {
		camTurnSpeed = CamTurnSpeedDps * TimeSecsElapsed
		CamRot.Y = math.Mod(CamRot.Y + ((CamTurnRight - CamTurnLeft) * camTurnSpeed), 360)
		CamRot.X = math.Mod(CamRot.X + ((CamTurnDown - CamTurnUp) * camTurnSpeed), 360)
		UpdateCamRot()
	}
	if CamMove {
		camMoveSpeed = CamMoveSpeedMps * TimeSecsElapsed
		CamPos.X += camMoveSpeed * (((CamMoveFwd - CamMoveBack) * CamSin.Y * math.Abs(CamCos.X)) - ((CamMoveLeft - CamMoveRight) * CamCos.Y))
		CamPos.Y += camMoveSpeed * ((-(CamMoveFwd - CamMoveBack) * CamSin.X) - ((CamMoveDown - CamMoveUp) * CamCos.X))
		CamPos.Z += camMoveSpeed * (((CamMoveFwd - CamMoveBack) * CamCos.Y * math.Abs(CamCos.X)) + ((CamMoveLeft - CamMoveRight) * CamSin.Y))
	}
	if CamTurn || CamMove {
		UpdateCamLook()
	}
}

func Reinit () {
	log.Printf("Pipeline init...")
	CleanUp(true)
	UpdateCamRot()
	UpdateCamLook()
}

func SpeedKmh () float64 {
	return (CamMoveSpeedMps * 3600) / 1000
}

func UpdateCamLook () {
// 			rline.x1 = (ln.z1 * math.Sin(rotr)) + (ln.x1 * math.Cos(rotr))
// 			rline.z1 = (ln.z1 * math.Cos(rotr)) - (ln.x1 * math.Sin(rotr))
// 			rline.x2 = (ln.z2 * math.Sin(rotr)) + (ln.x2 * math.Cos(rotr))
// 			rline.z2 = (ln.z2 * math.Cos(rotr)) - (ln.x2 * math.Sin(rotr))
	CamLook.X = CamPos.X + ((1 * CamSin.Y * math.Abs(CamCos.X)) - (0 * CamCos.Y))
	CamLook.Y = CamPos.Y + ((-1 * CamSin.X) - (0 * CamCos.X))
	CamLook.Z = CamPos.Z + ((1 * CamCos.Y * math.Abs(CamCos.X)) + (0 * CamSin.Y))
}

func UpdateCamRot () {
	CamRad.SetFromDegToRad(&CamRot)
	CamCos.SetFromCos(&CamRad)
	CamSin.SetFromSin(&CamRad)
}
