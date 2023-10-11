#version 330 core
in vec4  vColor;
in vec2  vTexCoords;
in float vIntensity;
in vec4  vClipRect;

out vec4 fragColor;

uniform vec4 uColorMask;
uniform vec4 uTexBounds;
uniform sampler2D uTexture;
uniform vec4 uClipRect;

void main() {
	if ((vClipRect != vec4(0,0,0,0)) && (gl_FragCoord.x < vClipRect.x || gl_FragCoord.y < vClipRect.y || gl_FragCoord.x > vClipRect.z || gl_FragCoord.y > vClipRect.w))
		discard;
	fragColor = vColor;
	if (vIntensity == 0) {
		fragColor *= vColor * texture(uTexture, vTexCoords).a;
		fragColor *= uColorMask;
	} else {
		fragColor *= vColor * texture(uTexture, vTexCoords);
		fragColor *= uColorMask;
	}
}
