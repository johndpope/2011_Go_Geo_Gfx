import numpy, wx
from numpy import arange, array, float32, int8

from VolRenderSkel import *

fragment_shader_src = '''
        uniform sampler1D TransferFunction;
        uniform sampler3D VolumeData;
        uniform sampler2D RayEnd;
        uniform sampler2D RayStart;
        
        float stepSize = 0.001; 
        
        void main(void)
        {
        
            // Get the end point of the ray (from the front-culled faces rendering)
            vec3 rayStart = texture2D(RayStart, gl_TexCoord[0].st).xyz;
            vec3 rayEnd = texture2D(RayEnd, gl_TexCoord[0].st).xyz;
            
            // Get a vector from back to front
            vec3 traverseVector = rayEnd - rayStart;

            // The maximum length of the ray
            float maxLength = length(traverseVector);
              
            // Construct a ray in the correct direction and of step length
            vec3 step = stepSize * normalize(traverseVector);
            vec3 ray = step;
       
            // The color accumulation buffer
            vec4 acc = vec4(0.0, 0.0, 0.0, 0.0);

            // Holds current voxel color
            vec4 voxelColor;

            // Advance ray
            for (int i = 0; i < int(1/stepSize); ++i)
            {
                if (length(ray) >= maxLength || acc.a >= 0.99) 
                {
                    acc.a = 1.0;
                    break;
                }

                voxelColor = texture1D(TransferFunction, texture3D(VolumeData, ray + rayStart).w);

                // Accumulate RGB : acc.rgb = voxelColor.rgb*voxelColor.a + (1.0 - voxelColor.a)*acc.rgb;
                acc.rgb = mix(acc.rgb, voxelColor.rgb, voxelColor.a);

                // Accumulate Opacity: acc.a = acc.a + (1.0 - acc.a)*voxelColor.a;
                acc.a = mix(voxelColor.a, 1.0, acc.a);

                ray += step;

            }
        
            gl_FragColor = acc;
            return;
        }
        '''

class RayCaster(VolumeRenderSkeleton):
       
    def __init__(self, parent):
        
        VolumeRenderSkeleton.__init__(self, parent)

        self.fragment_src_file = 'raycast.f.c'
        self.vertex_src_file = 'raycast.v.c'

        self.data_scale = array([1.0, 1.0, 1.0], dtype=float32)
        self.iso_value = 0.0

        # Shader sources
        self.vertex_color_shader = '''
                void main(void)
                {
                    gl_FrontColor = vec4(gl_Vertex.xyz, 1.0);
                    gl_Position = gl_ModelViewProjectionMatrix * gl_Vertex;
                }
                '''
        
        # Default
        self.fragment_shader_src = fragment_shader_src
        self.vertex_shader_src = '''
            void main(void)
            {
                gl_TexCoord[0] = gl_Vertex;
                gl_Position = gl_ModelViewProjectionMatrix * gl_Vertex;
            }
        '''


    def InitGL(self):
        VolumeRenderSkeleton.InitGL(self)
        self.color_program = compile_program(self.vertex_color_shader, None)        
        self.BuildGeometry()

        return
    
    def SetupUniforms(self):
        VolumeRenderSkeleton.SetupUniforms(self)

        glActiveTexture(GL_TEXTURE1)
        glBindTexture(GL_TEXTURE_3D, self.vol_data)
        glUniform1i(glGetUniformLocation(self.program, "VolumeData"), 1)
        glUniform1i(glGetUniformLocation(self.program, "RayEnd"), 2)
        glUniform1i(glGetUniformLocation(self.program, "RayStart"), 3)

    def OnDraw(self,event):
        self.SetCurrent()
        if not self.init:
            self.InitGL()
            self.init = True

        glClear(GL_COLOR_BUFFER_BIT | GL_DEPTH_BUFFER_BIT)
        glLoadIdentity()

        glTranslate(0.0, 0.0, -2.6)
        glPushMatrix()
        glRotate(self.rotation_y, 0.0, 1.0, 0.0)
        glRotate(self.rotation_x, 1.0, 0.0, 0.0)
        glTranslate(-0.5, -0.5, -0.5)

        # Render the back texture
        glEnable(GL_CULL_FACE)
        glCullFace(GL_FRONT)
        glUseProgram(self.color_program)
        glBindBuffer(GL_ARRAY_BUFFER, self.cube_vbo)
        glInterleavedArrays(GL_V3F, 0, None)
        glDrawArrays(GL_QUADS, 0, 4*8)
        glBindBuffer(GL_ARRAY_BUFFER, 0)

        # Get the texture data
        glActiveTexture(GL_TEXTURE2)
        glBindTexture(GL_TEXTURE_2D, self.back_texture)
        glCopyTexSubImage2D(GL_TEXTURE_2D, 0, 0, 0, 0, 0, 512, 512)

        # Render the front texture
        glClear(GL_COLOR_BUFFER_BIT | GL_DEPTH_BUFFER_BIT)
        glCullFace(GL_BACK)
        glBindBuffer(GL_ARRAY_BUFFER, self.cube_vbo)
        glInterleavedArrays(GL_V3F, 0, None)
        glDrawArrays(GL_QUADS, 0, 4*8)
        # Get the texture data
        glActiveTexture(GL_TEXTURE3)
        glBindTexture(GL_TEXTURE_2D, self.front_texture)
        glCopyTexSubImage2D(GL_TEXTURE_2D, 0, 0, 0, 0, 0, 512, 512)

        # Clear the screen
        glClear(GL_COLOR_BUFFER_BIT | GL_DEPTH_BUFFER_BIT)
        glDisable(GL_CULL_FACE)
        glBindBuffer(GL_ARRAY_BUFFER, 0)
        glUseProgram(None)

        # Render the wireframe box
        glDisableClientState(GL.GL_COLOR_ARRAY)
        glColor(0.0, 1.0, 0.0)
        glDisable(GL_LIGHTING)
        glDisable(GL_TEXTURE_2D)
        glVertexPointerf(self.box)
        glDrawArrays(GL_LINES, 0, len(self.box))
        glPopMatrix()
        
        # Render the cube
        glUseProgram(self.program)
        
        # Setup the texture information
        self.SetupUniforms()

        # Render flat quad and fudge it to the center 
        glTranslate(-0.58, -0.58, -0.5)
        glColor(0.0, 0.0, 1.0)
        glVertexPointerf(self.flat_plane)
        glDrawArrays(GL_QUADS, 0, 4)
        
        self.SwapBuffers()
        
        return

    def BuildGeometry(self):
        
        cube_panels = []

        for t in ['xy', 'xz', 'yz']:
            for p in [0.0, 1.0]:
                plane_verts = gen_plane(t=t, p=p)

                if not int(p):
                    plane_verts.reverse()

                for pv in plane_verts:
                    # Colors
                    cube_panels.extend(pv)
         
        self.cube_vbo = simple.GLuint(0)
        self.cube_data = numpy.array(cube_panels, dtype=numpy.float32)
        glGenBuffers(1, self.cube_vbo)
        glBindBuffer(GL_ARRAY_BUFFER, self.cube_vbo)
        glBufferData(GL_ARRAY_BUFFER, ADT.arrayByteCount(self.cube_data), 
                     ADT.voidDataPointer(self.cube_data), GL_STATIC_DRAW_ARB)
        
        self.flat_plane = numpy.array(gen_plane(t='xy', p=2.0));
        
        self.box = [[0.0, 0.0, 0.0],[0.0, 0.0, 1.0],
                    [1.0, 0.0, 0.0],[1.0, 0.0, 1.0],
                    [1.0, 1.0, 0.0],[1.0, 1.0, 1.0],
                    [0.0, 1.0, 0.0],[0.0, 1.0, 1.0]]
        
        self.box.extend(box_side())
        self.box.extend(box_side(z=1.0))

    def LoadVolumeData(self):

        # Load the volume data
        buffer = open('data/MRI_head.dat', 'rb').read()
        data = numpy.frombuffer(buffer, numpy.int8)
        self.vol_data = glGenTextures(1)
        glPixelStorei(GL_UNPACK_ALIGNMENT,1)
        glBindTexture(GL_TEXTURE_3D, self.vol_data)
        glTexParameterf(GL_TEXTURE_3D, GL_TEXTURE_WRAP_S, GL_CLAMP)
        glTexParameterf(GL_TEXTURE_3D, GL_TEXTURE_WRAP_T, GL_CLAMP)
        glTexParameterf(GL_TEXTURE_3D, GL_TEXTURE_WRAP_R, GL_CLAMP)
        glTexParameterf(GL_TEXTURE_3D, GL_TEXTURE_MAG_FILTER, GL_LINEAR)
        glTexParameterf(GL_TEXTURE_3D, GL_TEXTURE_MIN_FILTER, GL_LINEAR)
    
        glTexImage3D(GL_TEXTURE_3D, 0, GL_ALPHA, 128, 128, 128, 0, GL_ALPHA, GL_UNSIGNED_BYTE, data)
    
        # Create a texture for the ends of the rays
        self.back_texture = glGenTextures(1)
        glPixelStorei(GL_UNPACK_ALIGNMENT,1)
        glBindTexture(GL_TEXTURE_2D, self.back_texture)
        glTexParameterf(GL_TEXTURE_2D, GL_TEXTURE_WRAP_S, GL_CLAMP)
        glTexParameterf(GL_TEXTURE_2D, GL_TEXTURE_WRAP_T, GL_CLAMP)
        glTexParameterf(GL_TEXTURE_2D, GL_TEXTURE_MAG_FILTER, GL_LINEAR)
        glTexParameterf(GL_TEXTURE_2D, GL_TEXTURE_MIN_FILTER, GL_LINEAR)
    
        glTexImage2D(GL_TEXTURE_2D, 0, GL_RGB, 512, 512, 0, GL_RGB, GL_UNSIGNED_BYTE, None)
        
        # Create a texture for the start of the rays
        self.front_texture = glGenTextures(1)
        glPixelStorei(GL_UNPACK_ALIGNMENT,1)
        glBindTexture(GL_TEXTURE_2D, self.front_texture)
        glTexParameterf(GL_TEXTURE_2D, GL_TEXTURE_WRAP_S, GL_CLAMP)
        glTexParameterf(GL_TEXTURE_2D, GL_TEXTURE_WRAP_T, GL_CLAMP)
        glTexParameterf(GL_TEXTURE_2D, GL_TEXTURE_MAG_FILTER, GL_LINEAR)
        glTexParameterf(GL_TEXTURE_2D, GL_TEXTURE_MIN_FILTER, GL_LINEAR)
    
        glTexImage2D(GL_TEXTURE_2D, 0, GL_RGB, 512, 512, 0, GL_RGB, GL_UNSIGNED_BYTE, None)
        return

if __name__ == '__main__':
    app = wx.App()
    glutInit()
    frame = wx.Frame(None, -1, 'Volume RayCasting with PyOpenGL', wx.DefaultPosition, wx.Size(600, 600))
    canvas = RayCaster(frame)

    frame.Show()
    app.MainLoop()

    # Cleanup
