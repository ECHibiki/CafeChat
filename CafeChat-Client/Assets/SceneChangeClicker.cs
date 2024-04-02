using System.Collections;
using System.Collections.Generic;
using System.Linq.Expressions;
using UnityEngine;
using UnityEngine.UI;

public class SceneChangeClicker : MonoBehaviour
{
    // Start is called before the first frame update
    public string channelName;
    public SceneController sceneController;
    public Button button;
    public SpriteRenderer bg;
    public SpriteRenderer neutralMascot;
    public SpriteRenderer embarrasedMascot;
    public SpriteRenderer angryMascot;
    public SpriteRenderer smileMascot;
    public SpriteRenderer characterBubbleGraphic;
    public TMPro.TMP_Text characterBubbleText;
    
    public SpriteRenderer[] hiddingSprites;
    public GameObject mascotContainer;

    public ExpressionChanger neutralButton;    
    public ExpressionChanger embarrasedButton;
    public ExpressionChanger angryButton;
    public ExpressionChanger smileButton;
    public float transitionX = -227.4f;
    public float transitionY = 180.56f;

    void Start()
    {
        button.onClick.AddListener(OnButtonClicked);
    }
    void OnButtonClicked(){
        Debug.Log("Button clicked " + channelName);
        sceneController.OnButtonClicked(
            bg,
            mascotContainer,
            neutralMascot,
            characterBubbleGraphic,
            characterBubbleText,
            channelName,
            transitionX,
            transitionY,
            hiddingSprites
        );
        if (channelName != "home"){
            neutralButton.activatorExpression = neutralMascot;
            neutralButton.otherExpressions = new SpriteRenderer[]{embarrasedMascot, angryMascot, smileMascot};
            embarrasedButton.activatorExpression = embarrasedMascot;
            embarrasedButton.otherExpressions = new SpriteRenderer[]{neutralMascot, angryMascot, smileMascot};
            angryButton.activatorExpression = angryMascot;
            angryButton.otherExpressions = new SpriteRenderer[]{neutralMascot, embarrasedMascot, smileMascot};
            smileButton.activatorExpression = smileMascot;
            smileButton.otherExpressions = new SpriteRenderer[]{neutralMascot, embarrasedMascot, angryMascot};
        }

    }
}
