
using UnityEngine;

public class CameraAspect : MonoBehaviour
{

    public Camera cam;
    public float aspectRatio = 16f / 9f;
    // Start is called before the first frame update
    void Update()
    {   
        Adjust();
    }

    public void Adjust(){
        float windowAspect = (float)Screen.width / (float)Screen.height;
        float scaleHeight = windowAspect / aspectRatio;
        
        if(scaleHeight < 1f){
            Rect rect = cam.rect;
            rect.width = 1f;
            rect.height = scaleHeight;
            rect.x = 0;
            rect.y = (1f - scaleHeight) / 2f;
            cam.rect = rect;
        } else {
            float scaleWidth = 1f / scaleHeight;
            Rect rect = cam.rect;
            rect.width = scaleWidth;
            rect.height = 1f;
            rect.x = (1f - scaleWidth) / 2f;
            rect.y = 0;
            cam.rect = rect;
        }
    }
}
