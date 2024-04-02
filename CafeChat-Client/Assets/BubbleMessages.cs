using System.Collections;
using System.Collections.Generic;
using TMPro;
using UnityEngine;

public class BubbleMessages : MonoBehaviour
{
    // Start is called before the first frame update
    public SpriteRenderer adminBubbleGraphic;
    public TMP_Text adminBubbleText;
    public SpriteRenderer characterBubbleGraphic;
    public TMP_Text characterBubbleText;



    public void bubbleFadeIn(string message, bool isAdmin){
        Debug.Log("Fading in bubble");try{
            StopAllCoroutines();
        } catch {
            Debug.Log( "error " + "No coroutines to stop");
        }
        if (isAdmin){
            adminBubbleText.text = message;
            StartCoroutine(FadeIn(adminBubbleGraphic, adminBubbleText));
            characterBubbleGraphic.color = new Color(1, 1, 1, 0);
            characterBubbleText.color = new Color(0, 0, 0, 0);
            StartCoroutine(FadeOut(adminBubbleGraphic, adminBubbleText));
        } else {
            characterBubbleText.text = message;
            StartCoroutine(FadeIn(characterBubbleGraphic, characterBubbleText));
            adminBubbleGraphic.color = new Color(1, 1, 1, 0);
            adminBubbleText.color = new Color(0, 0, 0, 0);
            StartCoroutine(FadeOut(characterBubbleGraphic, characterBubbleText));
        }
    }

    IEnumerator FadeIn(SpriteRenderer bubble, TMP_Text text){
        // fade in
        yield return new WaitForSeconds(1f);
        Color cb = bubble.color;
        Color ct = text.color;
        for (float f = bubble.color.a; f <= 1.1; f += 0.1f){
            cb.a = f;
            ct.a = f;
            bubble.color = cb;
            text.color = ct;
            yield return new WaitForSeconds(0.05f);
        }
    }   
    IEnumerator FadeOut(SpriteRenderer bubble, TMP_Text text){
        yield return new WaitForSeconds(8f);
        Color cb = bubble.color;
        Color ct = text.color;
        for (float f = bubble.color.a; f >= -0.01f; f -= 0.05f){
            cb.a = f;
            ct.a = f;
            bubble.color = cb;
            text.color = ct;
            yield return new WaitForSeconds(0.05f);
        }
    }
}
