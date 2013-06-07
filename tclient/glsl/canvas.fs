uniform vec2 uScreen;
uniform sampler2D uTex0;

out vec4 vCol;

void main (void) {
	vCol = texture(uTex0, uScreen * gl_FragCoord.xy);
	vCol.a = 1;
}
