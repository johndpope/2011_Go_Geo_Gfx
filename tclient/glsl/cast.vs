uniform vec3 uCamLook;
uniform vec3 uCamPos;
uniform vec2 uScreen;

out vec3 vRayDir;

void main (void) {
	vec2 vPos = vec2((4 * (gl_VertexID % 2)) - 1, (4 * (gl_VertexID / 2)) - 1);
	// vPos.y = -vPos.y;
	gl_Position = vec4(vPos.xy, 0, 1);
	vPos = (vPos + 1) * 0.5;

	const vec2 vViewPlane = ((vPos * 2) - 1) / vec2(1, uScreen.y / uScreen.x);
	const vec3 vForwards = normalize(uCamLook - uCamPos);
	const vec3 vRight = normalize(cross(vForwards, vec3(0, 1, 0)));
	const vec3 vUp = cross(vRight, vForwards);
	vRayDir = (-vRight * vViewPlane.x) + (vUp * vViewPlane.y) + vForwards;
}
