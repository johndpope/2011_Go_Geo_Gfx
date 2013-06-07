uniform vec4 uScreen;
uniform sampler2D uTex0;

out vec3 vLum;

void main (void) {
	const vec2 vOffset = vec2(-0.5, 0.5);
	const vec2 vPos = gl_FragCoord.xy * uScreen.xy;
	float fGray;
	float fAvg = 0;
	float fMax = -1e20;
	float fMin = 1e19;

	for (int x = 0; x < 2; x++) {
		for (int y = 0; y < 2; y++) {
			fGray = dot(texture(uTex0, vPos + vec2(vOffset[x] * uScreen[2], vOffset[y] * uScreen[3])).rgb, vec3(0.299, 0.587, 0.114));
			fAvg += log(0.0001 + fGray);
			fMax = max(fMax, fGray);
			fMin = min(fMin, fGray);
		}
	}
	vLum = vec3(fAvg / 4, fMax, fMin);
}
