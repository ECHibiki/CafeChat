using System.Collections;
using System.Collections.Generic;
using System.Threading;
using TMPro;
using Unity.VisualScripting;
using UnityEngine;
using UnityEngine.UI;
public class SceneController : MonoBehaviour
{    
    
    public Connection connection;

    public GameObject roomButtonContainer;
    public GameObject homeButton;
    public GameObject expressionButtonContainer;

    public BubbleMessages bubbleMessages;
    
    
    public void OnButtonClicked(SpriteRenderer boardbg, GameObject boardmascotContainer, SpriteRenderer boardmascot,
            SpriteRenderer characterBubbleGraphic, TMP_Text characterBubbleText,
         string channel, float leftMove, float upMove, SpriteRenderer[] hiddingSprites){
        // alpha decay ove rone seconds on homeBG and fade in on qaBG
        Debug.Log("Button clicked scene transition");
        
        Color c = bubbleMessages.characterBubbleGraphic.color;
        c.a = 0;
        bubbleMessages.characterBubbleGraphic.color = c;
        c = bubbleMessages.characterBubbleText.color;
        c.a = 0;
        bubbleMessages.characterBubbleText.color = c;
        
        if (characterBubbleGraphic != null && characterBubbleText != null){
            bubbleMessages.characterBubbleGraphic = characterBubbleGraphic;
            bubbleMessages.characterBubbleText = characterBubbleText;
        }
        
        boardbg.enabled = true;
        bool isHome = channel == "home";

        StartCoroutine(ManageButtons(isHome));

        foreach(SpriteRenderer sr in hiddingSprites){
            StartCoroutine(FadeOut(sr));
        }

        StartCoroutine(FadeIn(boardbg));
        StartCoroutine(FadeIn(boardmascot));
        StartCoroutine(MoveLeft(boardmascotContainer, leftMove));
        StartCoroutine(MoveUp(boardmascotContainer, upMove));


        connection.activeBoard = channel;
        connection.SendWebSocketMessage("/join #" + channel);
    }  
  
    IEnumerator ManageButtons(bool isHome){
        roomButtonContainer.SetActive(false);
        homeButton.SetActive(false);
        expressionButtonContainer.SetActive(true);
        yield return new WaitForSeconds(1f);
        if (isHome){
            roomButtonContainer.SetActive(true);
        }
        // expressionButtonContainer.SetActive(!isHome);
        homeButton.SetActive(!isHome);
    }

    IEnumerator FadeIn(SpriteRenderer sr){
        // fade in
        for (float f = 0.01f; f <= 1.1; f += 0.04f){
            Color c = sr.color;
            c.a = f;
            sr.color = c;
            yield return new WaitForSeconds(0.05f);
        }
    }
    IEnumerator FadeOut(SpriteRenderer sr){
        // fade out
        Color c = sr.color;
        for (float f = c.a; f >= -0.05f; f -= 0.05f){
            c.a = f;
            sr.color = c;
            yield return new WaitForSeconds(0.05f);
        }
    }

    IEnumerator MoveLeft(GameObject sr, float x){
        // move left
        Debug.Log("Moving left" + x + " " + sr.transform.localPosition.x);
        for (float f = sr.transform.localPosition.x; f >= x; f -= 5f){
            Vector3 v = sr.transform.localPosition;
            v.x = f;
            sr.transform.localPosition = v;
            yield return new WaitForSeconds(0.05f);
        }
    }

    IEnumerator MoveUp(GameObject sr, float y){
        // move up
        for (float f = sr.transform.localPosition.y; f <= y; f += 5f){
            Vector3 v = sr.transform.localPosition;
            v.y = f;
            sr.transform.localPosition = v;
            yield return new WaitForSeconds(0.05f);
        }
    }

}
