uniform vec4 uScreen;
uniform sampler2D uTex0;

out vec3 vLum;

void main (void) {
	const vec3 vOffset = vec3(-0.5, 0, 0.5);
	const vec2 vPos = gl_FragCoord.xy * uScreen.xy;
	bool bLast = (uScreen[2] == 3);
	vec2 vOff;
	vec3 vMidMax;
	float fAvg = 0;
	float fMax = -1e20;
	float fMin = 1e19;

	for (int x = 0; x < 3; x++) {
		for (int y = 0; y < 3; y++) {
			if (bLast) {
				vOff = vec2(x * 0.49999998, y * 0.49999998);
				vMidMax = texture(uTex0, vOff).rgb;
			} else {
				vOff = vec2(vOffset[x] * uScreen[2], vOffset[y] * uScreen[3]);
				vMidMax = texture(uTex0, vPos + vOff).rgb;
			}
			fAvg += vMidMax.r;
			fMax = max(fMax, vMidMax.g);
			fMin = min(fMin, vMidMax.b);
		}
	}
	vLum = vec3(fAvg / 9, fMax, fMin);
}
