uniform vec2 uScreen;
uniform sampler2D uTex0;

out vec3 vCol;

void main (void) {
	vCol = texture(uTex0, gl_FragCoord.xy * uScreen).rgb;
}
