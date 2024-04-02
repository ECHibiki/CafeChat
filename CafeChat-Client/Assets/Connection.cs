using System;
using System.Collections;
using System.Collections.Generic;
using UnityEngine;

using NativeWebSocket;

using UnityEngine.UI;
using TMPro;
using UnityEngine.TextCore.Text;
using System.Runtime.CompilerServices;
using System.Threading;
using System.Text.RegularExpressions;
using System.Linq;

public class Connection : MonoBehaviour
{
  WebSocket websocket;
  public TMP_InputField ChatText;

  public Scrollbar scrollbar;
  public BubbleMessages bubbleMessages;
  string textdata;
  
  public TMP_Text errorText;
  public TMP_Text closeText;

  public Button angryButton;
  public Button neutralButton;
  public Button embarrassedButton;
  public Button smileButton;
  public string activeBoard = "home";

  bool midConnectionAttempt = false;

  // Start is called before the first frame update
  void Start()
  {
    ConnectWebSocket();
  }

  IEnumerator Reconect(){
    if (midConnectionAttempt){
      Debug.Log("Already in the middle of a connection attempt");
      yield break;
    } else{
        ConnectWebSocket();
    }
  }

  async void ConnectWebSocket()
  {
    midConnectionAttempt = true;
    websocket = new WebSocket("wss://cafechat.kissu.moe/ws?channel=" + activeBoard);

    websocket.OnOpen += () =>
    {
      Debug.Log("Connection open!");
      scrollbar.value = 1;
      midConnectionAttempt = false;
    };
    websocket.OnClose += (e) =>
    {
      Debug.Log("Connection closed!");
      
      closeText.text = "Cl: " + e;

      float prevScroll = scrollbar.value;
      // getting the message as a string
      textdata += "Connection closed! Retrying in 10s..." + "\n";
      ChatText.text = textdata;
      if(prevScroll == 1){
        scrollbar.value = 1;
        Debug.Log("Scrolling to bottom" +  scrollbar.value);
      }
      websocket = null;
      midConnectionAttempt = false;
      StartCoroutine(Reconect());
      midConnectionAttempt = true;
    };

    websocket.OnError += (e) =>
    {
      errorText.text = "err: " + e;
      Debug.Log("Error! " + e);
      // StartCoroutine(Reconect());
    };

    websocket.OnMessage += (bytes) =>
    {
      float prevScroll = scrollbar.value;
      // getting the message as a string
      var message = System.Text.Encoding.UTF8.GetString(bytes);
      Debug.Log("OnMessage! " + message + " " + prevScroll);

      string character = message.Substring(0, message.IndexOf(":")); 
      if (character.IndexOf("System") != 0 && character != "Anonymous"){
        Debug.Log("Character: " + character);        
        if (character.IndexOf("verm-sama") == 0){
          // admin bubble
          message = message.Replace("||", "");
          bubbleMessages.bubbleFadeIn(message.Substring(message.IndexOf(":")+1),  true);
        } else {
          // character bubble
            // split character by || to exctract name and emotions
            string name = character.Substring(0, character.IndexOf("||"));
            string emotionsChk = character.Substring(character.IndexOf("||")+2);
            string[] emotions = new string[]{}; 
            if (emotionsChk.Length == 0){
              // no emotions detected
              Debug.Log("No emotion str");
            } else{
              Debug.Log("Emotion str" + emotionsChk +".");
              emotions = emotionsChk.Split(",");
            }
            string[][] emotionsDetails = new string[emotions.Length][];
            for (int i = 0; i < emotions.Length; i++){
              emotionsDetails[i] = emotions[i].Split("&");
              Debug.Log("EMO" + emotionsDetails[i]);
            }
            Debug.Log(emotions + " " +  emotions.Length);
            // Trigger emotion change
            string maxEmotion = MapExpressions( name , emotionsDetails);
            Debug.Log(maxEmotion);
            if(emotions.Length == 0){
              Debug.Log("No emotions detected");
            } else if (maxEmotion == "angry"){
              angryButton.onClick.Invoke();
            } else if (maxEmotion == "embarrassed"){
              embarrassedButton.onClick.Invoke();
            } else if (maxEmotion == "smile"){
              smileButton.onClick.Invoke(); 
            }else {
              neutralButton.onClick.Invoke();
            }

            // Send only the speach components of message body
            // use regex to only get text within quotes
            string messageText = message.Substring(message.IndexOf(":")+1);
            try{
                MatchCollection onlyQuotes = Regex.Matches(messageText, "\"([^\"]*?)\"");
                string[] onlyQuotesArr = new string[ onlyQuotes.Count ];
                Debug.Log("Only quotes" + onlyQuotes.Count);
                for (int i = 0; i < onlyQuotes.Count; i++){
                  onlyQuotesArr[i] = "'" + onlyQuotes[i].Groups[1].Value + "'";
                }
                string speach = string.Join("\n", onlyQuotesArr);
                bubbleMessages.bubbleFadeIn(speach, false);
                // alter message to only contain the sender and body for writting and storage
             } catch (Exception e){
              Debug.Log("Error on \" parse: " + e);  
            } finally{
              messageText = Regex.Replace(messageText, "\"([^\"]*?)\"", "<color=#9C4343>'$1'</color>");
            }

            message = name + ":" + messageText;         
        }
      }

      textdata += message + "\n";
      string[] lines = textdata.Split('\n');
      if (lines.Length > 15){
        //remove the first lines item
        lines = lines.Skip(1).ToArray();
      }
      try{
        ChatText.text = textdata;
      } catch (Exception e){
        Debug.Log("Error on ChatText.text = textdata: " + e);
        //remove the first lines item
        lines = lines.Skip(1).ToArray();
        ChatText.text = lines.Aggregate((a, b) => a + "\n" + b);
      }
      if(prevScroll == 1){
        scrollbar.value = 1;
        Debug.Log("Scrolling to bottom" +  scrollbar.value);
      }
    };

    // // Keep sending messages at every 0.3s
    // InvokeRepeating("PingWebSocket", 0.0f, 10f);

    // waiting for messages
    await websocket.Connect();
  }

    /*
      qa-chan
      angry:
      anger, annoyance, disappointment, disapproval, disgust

      smile:
      admiration, amusement, approval, caring, desire, excitement, gratitude, joy, optimism, pride

      neutral:
      neutral, realization, remorse, sadness

      blushing:
      surprise, confusion, curiosity, embarrassment, nervousness, surprise, relief, love, fear, grief
      */
      /*
      jp-chan
      angry:
      anger, annoyance, confusion, disappointment, disapproval, disgust, fear, grief, 

      smile:
      admiration, amusement, approval, caring, desire, excitement, gratitude, joy, love, optimism, pride, realization, relief

      neutral:
      grief, neutral, realization, remorse, sadness, surprise, confusion, curiosity, embarrassment, nervousness, surprise
    */
  string MapExpressions(string name, string[][] emotionsDetails){
    string maxEmotion = "neutral";
    float maxEmotionValue = 0;
     switch (name){
        case "qa-chan":
          for (int i = 0; i < emotionsDetails.Length; i++){
            switch (emotionsDetails[i][0]){
              case "anger":
              case "annoyance":
              case "disappointment":
              case "disapproval":
              case "disgust":
                if (float.Parse(emotionsDetails[i][1]) > maxEmotionValue){
                  maxEmotion = "angry";
                  maxEmotionValue = float.Parse(emotionsDetails[i][1]);
                }
                break;
              case "admiration":
              case "amusement":
              case "approval":
              case "caring":
              case "desire":
              case "excitement":
              case "gratitude":
              case "joy":
              case "optimism":
              case "pride":
                if (float.Parse(emotionsDetails[i][1]) > maxEmotionValue){
                  maxEmotion = "smile";
                  maxEmotionValue = float.Parse(emotionsDetails[i][1]);
                }
                break;
              case "neutral":
              case "realization":
              case "remorse":
              case "sadness":
                if (float.Parse(emotionsDetails[i][1]) > maxEmotionValue){
                  maxEmotion = "neutral";
                  maxEmotionValue = float.Parse(emotionsDetails[i][1]);
                }                    
                break;
                //   surprise, confusion, curiosity, embarrassment, nervousness, surprise, relief, love, fear, grief
              case "surprise":
              case "confusion":
              case "curiosity":
              case "nervousness":
              case "relief":
              case "love":
              case "fear":
              case "grief":
                if (float.Parse(emotionsDetails[i][1]) > maxEmotionValue){
                  maxEmotion = "embarrassed";
                  maxEmotionValue = float.Parse(emotionsDetails[i][1]);
                }
                break;
            } 
          }
          break;
        case "jp-chan":
          for (int i = 0; i < emotionsDetails.Length; i++){
            Debug.Log(emotionsDetails[i][0] + " " + emotionsDetails[i][1]);
            switch (emotionsDetails[i][0]){
              case "anger":
              case "annoyance":
              case "confusion":
              case "disappointment":
              case "disapproval":
              case "disgust":
              case "fear":
              case "grief":
                if (float.Parse(emotionsDetails[i][1]) > maxEmotionValue){
                  maxEmotion = "angry";
                  maxEmotionValue = float.Parse(emotionsDetails[i][1]);
                }
                break;
              case "admiration":
              case "amusement":
              case "approval":
              case "caring":
              case "desire":
              case "excitement":
              case "gratitude":
              case "joy":
              case "love":
              case "optimism":
              case "pride":
              case "realization":
              case "relief":
                if (float.Parse(emotionsDetails[i][1]) > maxEmotionValue){
                  maxEmotion = "embarrassed";
                  maxEmotionValue = float.Parse(emotionsDetails[i][1]);
                }
                break;
              case "neutral":
              case "remorse":
              case "sadness":
              case "surprise":
              case "curiosity":
              case "embarrassment":
              case "nervousness":
                if (float.Parse(emotionsDetails[i][1]) > maxEmotionValue){
                  maxEmotion = "neutral";
                  maxEmotionValue = float.Parse(emotionsDetails[i][1]);
                }                    
                break;
            } 
          }
          break;
        case "ec-chan":
          for (int i = 0; i < emotionsDetails.Length; i++){
            Debug.Log(emotionsDetails[i][0] + " " + emotionsDetails[i][1]);
            switch (emotionsDetails[i][0]){
              case "anger":
              case "annoyance":
              case "confusion":
              case "disappointment":
              case "disapproval":
              case "disgust":
              case "fear":
              case "grief":              
              case "remorse":
              case "sadness":
                if (float.Parse(emotionsDetails[i][1]) > maxEmotionValue){
                  maxEmotion = "angry";
                  maxEmotionValue = float.Parse(emotionsDetails[i][1]);
                }
                break;
              case "admiration":
              case "amusement":
              case "approval":
              case "caring":
              case "desire":
              case "excitement":
              case "joy":
              case "love":
              case "optimism":
              case "pride":
              case "relief":
              case "embarrassment":
                if (float.Parse(emotionsDetails[i][1]) > maxEmotionValue){
                  maxEmotion = "embarrassed";
                  maxEmotionValue = float.Parse(emotionsDetails[i][1]);
                }
                break;
              case "neutral":
              case "realization":
              case "surprise":
              case "curiosity":
              case "nervousness":
              case "gratitude":
                if (float.Parse(emotionsDetails[i][1]) > maxEmotionValue){
                  maxEmotion = "neutral";
                  maxEmotionValue = float.Parse(emotionsDetails[i][1]);
                }                    
                break;
            }
          }
          break;
          default:
            Debug.Log("Character not found");
            break; 
      }
      Debug.Log(maxEmotion + " " + maxEmotionValue);
      return maxEmotion;
  }

  void Update()
  {
    #if !UNITY_WEBGL || UNITY_EDITOR
      if (websocket != null){
        websocket.DispatchMessageQueue();
      }
    #endif
    // websocket.DispatchMessageQueue();
  }
  public async void SendWebSocketMessage(string msg = "")
  {
    if (websocket.State == WebSocketState.Open)
    {
      await websocket.SendText(msg);
    }
  }

  private async void OnApplicationQuit()
  {
    await websocket.Close();
  }

}
