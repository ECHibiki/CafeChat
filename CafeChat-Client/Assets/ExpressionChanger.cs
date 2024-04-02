using System.Collections;
using System.Collections.Generic;
using System.Linq;
using UnityEngine;
using UnityEngine.UI;

public class ExpressionChanger : MonoBehaviour
{

    public SpriteRenderer activatorExpression;
    public SpriteRenderer[] otherExpressions;

    public Button button;

    // Start is called before the first frame update
    void Start()
    {
        button.onClick.AddListener(OnButtonClicked);
    }

    // Update is called once per frame
    void OnButtonClicked(){
        if( activatorExpression == null || otherExpressions == null || button == null){
            Debug.Log("ExpressionChanger not set up correctly");
            return;
        }
        Debug.Log("Button clicked");
        
        StartCoroutine(FadeIn(activatorExpression));
        StartCoroutine(FadeOut(otherExpressions));    
    }

  IEnumerator FadeIn(SpriteRenderer sr){
        // fade in
        Color c = sr.color;
        for (float f = sr.color.a; f <= 1.1; f += 0.2f){
            c.a = f;
            sr.color = c;
            yield return new WaitForSeconds(0.05f);
        }
    }
    IEnumerator FadeOut(SpriteRenderer[] sr){
        // fade out
        yield return new WaitForSeconds(0.15f);
        foreach(SpriteRenderer s in sr){
            Color c = s.color;
            for (float f = s.color.a; f >= -0.01f; f -= 0.05f){
                c.a = f;
                s.color = c;
                yield return new WaitForSeconds(0.05f);
            }
        }
    }
}
