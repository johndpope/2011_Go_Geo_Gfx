void main (void) {
	//const vec2 aPos = vec2((4 * (gl_VertexID % 2)) - 1, (4 * (gl_VertexID / 2)) - 1);
	//vPos = (aPos + 1) / 2;
	// vPos.y = -vPos.y;
	gl_Position = vec4((4 * (gl_VertexID % 2)) - 1, (4 * (gl_VertexID / 2)) - 1, 0, 1);
}
