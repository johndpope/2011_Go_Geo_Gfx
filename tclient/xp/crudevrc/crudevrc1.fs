uniform sampler1D uTexTransferFunc;
uniform sampler3D uTexVolume;
uniform sampler2D uTexEnd;
uniform sampler2D uTexStart;
varying vec2 vPos;

float stepSize = 0.001; 

void main (void) {
	// Get the end point of the ray (from the front-culled faces rendering)
	vec3 rayStart = texture2D(uTexStart, vPos).xyz;
	vec3 rayEnd = texture2D(uTexEnd, vPos).xyz;

	// Get a vector from back to front
	vec3 traverseVector = rayEnd - rayStart;

	// The maximum length of the ray
	float maxLength = length(traverseVector);

	// Construct a ray in the correct direction and of step length
	vec3 step = stepSize * normalize(traverseVector);
	vec3 rayStep = step;

	// The color accumulation buffer
	vec4 acc = vec4(0.0, 0.0, 0.0, 0.0);

	// Holds current voxel color
	vec4 voxelColor;

	// Advance ray
	for (int i = 0; i < int(1 / stepSize); ++i) {
		if ((length(rayStep) >= maxLength) || (acc.a >= 0.99)) {
			acc.a = 1.0;
			break;
		}

		voxelColor = texture1D(uTexTransferFunc, texture3D(uTexVolume, rayStep + rayStart).w);

		// Accumulate RGB : acc.rgb = (voxelColor.rgb * voxelColor.a) + ((1.0 - voxelColor.a) * acc.rgb);
		acc.rgb = mix(acc.rgb, voxelColor.rgb, voxelColor.a);

		// Accumulate Opacity: acc.a = acc.a + (1.0 - acc.a)*voxelColor.a;
		acc.a = mix(voxelColor.a, 1.0, acc.a);

		rayStep += step;
	}
	gl_FragColor = acc;
	return;
}
