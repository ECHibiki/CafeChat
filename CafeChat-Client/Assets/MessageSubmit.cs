using System.Collections;
using System.Collections.Generic;
using UnityEngine;
using TMPro;
using UnityEngine.EventSystems;
public class MessageSubmit : MonoBehaviour
{
    
    public TMP_InputField  mainInputField;
    public Connection connection;
    // Start is called before the first frame update
    void Start()
    {
        mainInputField.onSubmit.AddListener(onclickChatMessage);
        EventSystem.current.SetSelectedGameObject(
            mainInputField.gameObject
        );
    }

    public void onclickChatMessage(string msg){
        sendChatMessage(msg);
    }
    public void sendChatMessage(string msg, bool bypass = false){
        if(bypass || Input.GetKeyDown(KeyCode.Return) || 
            Input.GetKeyDown(KeyCode.KeypadEnter)){
                
            connection.SendWebSocketMessage(msg);
            
            EventSystem.current.SetSelectedGameObject(
                null
            );
            Debug.Log("end edit: " + mainInputField.text);
            mainInputField.text = "";
            EventSystem.current.SetSelectedGameObject(
                mainInputField.gameObject
            );
            StartCoroutine(activateInputField());
        } else{
            Debug.Log("invalid submit " + Input.anyKey);
        }
    }

    IEnumerator activateInputField(){
        yield return new WaitForSeconds(1f);
        EventSystem.current.SetSelectedGameObject(
            mainInputField.gameObject
        );
    }
}
