uniform vec2 uScreen;
uniform sampler2D uTex0;
uniform sampler2D uTex1;

out vec3 vCol;

const vec3 VNULL = vec3(0);

void Bleach () {
	const float fBleachAmount = 2;
	const vec3 vCoeff = vec3(0.2125, 0.7154, 0.0721);
	const float fLum = max(0.001, dot(vCol.rgb, vCoeff));
	const vec3 vLum = vec3(fLum);
	const float fMix = clamp((fLum - 0.4) * 10, 0, 1);
	const vec3 vB1 = 2 * vCol * vLum;
	const vec3 vB2 = 1 - (2 * (1 - vCol) * (1 - vLum));
	const vec3 vMix = mix(vB1, vB2, fMix);
	vCol = mix(vCol, vMix, fBleachAmount);
}

void BlueShift () {
	vCol = mix(vCol, vCol * vec3(1.05, 0.97, 1.27), 0.33);
}

void Colorful () {
	vCol *= step(0.05, vCol);
}

void Defog () {
	const vec3 vDefog = vec3(0);
	const vec3 vColor = vec3(0);
	vCol = max(VNULL, vCol - (vDefog * vColor));
}

float log10 (const in float fVal) {
	return log2(fVal) * log2(10);
}

void Exposure (/*const in vec3 vLum*/) {
	return;
//	const float fLumPixel = max(0.0001, dot(vCol, vec3(0.2126, 0.7152, 0.0722/*0.299, 0.587, 0.114*/)));
//	const float fMean = max(0.0001, exp(vLum.r + 0.0001));
//	const float fMidGrey = max(0, 1.5 - (1.5 / ((fMean * 0.1) + 1))) + 0.1;
//	const float fLumScaled = fMidGrey * (fLumPixel / fMean);
//	const float fMidGrey = 1.03 - (2 / (2 + log10(fMean + 1)));
//	const float fLumScaled = (fLumPixel * fMidGrey) / fMean;
//	const float fLumMapped = fLumScaled / (1 + fLumScaled);

//	const float fLin = fMidGrey / fMean;
//	const float fExposure = (max(fLin, 0.0001));
//	vCol *= exp2(fExposure);
}

void Grayscale () {
	vCol = vec3(dot(vCol, vec3(0.222, 0.707, 0.071)));
	// vCol = vec3((vCol.r + vCol.g + vCol.b) / 3);
}

vec3 Tonemap (const in vec3 vColor) {
	const float fShoulder = 0.33; // 0.22 // 0.15;
	const float fLinStrength = 0.2; // 0.30 // 0.50;
	const float fLinAngle = 0.1; // 0.10 // 0.10
	const float fToeStrength = 0.20; // 0.20 // 0.20
	const float fToeNumerator = 0.001; // 0.01 // 0.02;
	const float fToeDenominator = 0.20; // 0.30 // 0.33
	return ((vColor * ((fShoulder * vColor) + (fLinAngle * fLinStrength)) + (fToeStrength * fToeNumerator)) / (vColor * ((fShoulder * vColor) + fLinStrength) + (fToeStrength * fToeDenominator))) - (fToeNumerator / fToeDenominator);
}

void Tonemap () {
	const vec3 vNumerator = Tonemap(vCol * 1);
	const vec3 vDenominator = Tonemap(vec3(11.2));
	vCol = vNumerator / vDenominator;
}

void Vignette () {
	const float fAmount = -2; // -2 darkest, >0 whitening
	const float fRadius = 0.88;
	const vec2 vCenter = vec2(0.5);
	// vCol += (fAmount * pow(length(vPos - vCenter) / fRadius, 4));
}

void main () {
//	const float fBloom = 0.4;
	const vec2 vPos = uScreen * gl_FragCoord.xy;
//	const vec3 vBloom = texture(uTex1, vPos).rgb;
//	const vec3 vLum = texture(uTex1, vec2(0.5, 0.5)).rgb;
	vCol = texture(uTex0, vPos).rgb;
	//const float fFragLum = dot(vCol.rgb, vec3(0.299, 0.587, 0.114));
//	vCol = mix(vCol, vBloom * 1, fBloom);
	//if (vPos.x > 0.25)
	Exposure();
	//if (vPos.x > 0.45)
	Tonemap();
/*
	if (vPos.y < 0.05)
		vCol = vec3(exp(vLum.r));
	else if (vPos.y < 0.1)
		vCol = vec3(fExp / 10);
*/
}
