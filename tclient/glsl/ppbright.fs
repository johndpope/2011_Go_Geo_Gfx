uniform vec2 uScreen;
uniform sampler2D uTex0;
uniform sampler2D uTex1;

out vec3 vCol;

void main (void) {
	vCol = texture(uTex0, gl_FragCoord.xy * uScreen).rgb;
	vCol *= step(1, dot(vCol.rgb, vec3(0.299, 0.587, 0.114)));
	// vCol = vTexCol.rgb * step(vec3(1), vTexCol.rgb);
	/*
	if (vTexCol.r > 1 || vTexCol.g > 1 || vTexCol.b > 1) {
		vCol = vTexCol.rgb;
	} else {
		vCol = vec3(0);
	}
	/*
	const float fExposure = 1;
	const float fThreshold = 0.5;
	const vec4 vTexCol = texture(uTex0, vPos);
	const vec2 vLum = texture(uTex1, vec2(0.5, 0.5)).rg;
	const float fScaleLum = (vTexCol.a * fExposure) / vLum.r;
	vCol = max(vec3(0), (vTexCol.rgb * ((fScaleLum * (1 + (fScaleLum / (vLum.g * vLum.g)))) / (1 + fScaleLum))) - fThreshold);
	*/
}
