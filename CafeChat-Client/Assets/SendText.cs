
using UnityEngine;
using UnityEngine.UI;
 
public class SendText : MonoBehaviour
{
    // Start is called before the first frame update
    public MessageSubmit inputField;
    void Start()
    {
        // GetComponent<Button>().onClick.AddListener(sendText);
        GetComponent<Button>().onClick.AddListener(sendText);
    }

    public void sendText(){
        inputField.sendChatMessage(inputField.mainInputField.text, true);
    }

}
