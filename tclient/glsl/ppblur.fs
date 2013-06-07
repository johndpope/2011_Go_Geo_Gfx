in vec2 vPos;

uniform vec2 uScreen;
uniform sampler2D uTex0;

uniform float[] fOffsets = float[] (0, 1.3846153846, 3.2307692308);
uniform float[] fWeights = float[] (0.2270270270, 0.3162162162, 0.0702702703);

out vec3 vCol;

void main (void) {
	vCol = texture(uTex0, vPos).rgb * fWeights[0];
	for (int i = 1; i < 3; i++) {
		vCol += (texture(uTex0, vPos + vec2(fOffsets[i] * uScreen.x, fOffsets[i] * uScreen.y)).rgb * fWeights[i]);
		vCol += (texture(uTex0, vPos - vec2(fOffsets[i] * uScreen.x, fOffsets[i] * uScreen.y)).rgb * fWeights[i]);
	}
}
