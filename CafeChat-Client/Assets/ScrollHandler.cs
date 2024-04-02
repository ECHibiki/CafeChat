using System.Collections;
using System.Collections.Generic;
using UnityEngine;
using UnityEngine.UI;

public class ScrollHandler : MonoBehaviour
{
    // Start is called before the first frame update
    public Scrollbar scrollbar;
    public float scrollSensitive = 0.1f;
    void Start()
    {
        
    }

    // Update is called once per frame
    void Update()
    {
        float y =  Input.mouseScrollDelta.y * scrollSensitive;
        if (y != 0){
            Debug.Log(y);
            scrollbar.value += y;
            if (scrollbar.value < 0){
                scrollbar.value = 0;
            } else if (scrollbar.value > 1){
                scrollbar.value = 1;
            }
        }
    }
}
