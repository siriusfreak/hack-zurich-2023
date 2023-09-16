import React, { useState, useEffect } from 'react';
import { Container, TextField, Button, List, ListItem, ListItemText, Divider, MenuItem, Menu, IconButton, Box } from '@mui/material';
import SendIcon from '@mui/icons-material/Send';
import sikaLogo from './assets/LogoSika.png';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import Flag from 'react-country-flag';
import AddIcon from '@mui/icons-material/Add';
import Fab from '@mui/material/Fab';

const styles = {
  chatContainer: {
    padding: '50px',
    borderRadius: '10px',
    background: '#ffffff',
    height: '100vh',
    display: 'flex',
    flexDirection: 'column',
    justifyContent: 'space-between',
    margin: '0 auto',
    maxWidth: '800px',
  },
  chatHeader: {
    backgroundColor: '#ffffff',
    borderBottom: '1px solid #ccc',
    padding: '20px 0 0 0',
    fontSize: '20px',
    fontWeight: 'bold',
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
  },  
  sikaLogoImage: {
    width: '50px',
    height: 'auto',
  },
  messageList: {
    display: 'flex',
    flexDirection: 'column',
    flexGrow: 1,
    overflow: 'auto',
  },
  messageBubble: (isOwnMessage) => ({
    alignSelf: isOwnMessage ? 'flex-end' : 'flex-start',
    maxWidth: '50%',
    padding: '10px 20px',
    borderRadius: '20px',
    margin: '10px',
    backgroundColor: '#f5f5f5',
    textAlign: isOwnMessage ? 'right' : 'left',
  }),
  inputContainer: {
    display: 'flex',
    padding: '20px',
    backgroundColor: '#f0f0f0',
    borderRadius: '10px',
    position: 'relative',
  },
  inputField: {
    flexGrow: 1,
    borderRadius: '10px',
    overflow: 'hidden',
    marginRight: '10px',
    padding: '10px 20px 20px 0',
  },
  sendButton: {
    position: 'absolute',
    right: '10px',
    top: '50%',
    transform: 'translateY(-50%)',
    borderRadius: '50%',
    height: '40px',
    width: '40px',
    boxShadow: 'none',
    '&:hover': {
      backgroundColor: 'rgba(0, 0, 0, 0.04)',
    },
  },
  languageSelector: {
    display: 'flex',
    alignItems: 'center',
  },

  chatLayout: {
    backgroundColor: '#ffffff',
    borderBottom: '1px solid #ccc',
    display: 'flex',
    height: '100vh',
    borderRadius: '10px',

  },
  chatListContainer: {
    width: '300px',
    backgroundColor: '#f5f5f5',
    height: '100%',
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'flex-start',
    padding: '20px 0 0 0',
    boxSizing: 'border-box',
  },
  chatListHeader: {
    width: '100%',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    boxSizing: 'border-box',
    padding: '0 0 20px 0',
  },
  newChatButton: {
    backgroundColor: 'red',
    borderRadius: '50%',
    width: '40px',
    height: '40px',
    display: 'flex',
    justifyContent: 'center',
    alignItems: 'center',
    color: 'white',
    fontSize: '20px',
    cursor: 'pointer',
    padding: 0,
  },
  chatList: {
    width: '100%',
    background: '#f5f5f5',
    overflow: 'auto',
    maxHeight: 'calc(100vh - 120px)',
    color: 'white',
    backgroundColor: '#616161',
    borderRadius: '10px',
  },
  chatWindow: {
    flexGrow: 1,
    display: 'flex',
    flexDirection: 'column',
    background: '#ffffff',
    borderRadius: '10px',
    overflow: 'hidden',
  },
  activeChatItem: {
    backgroundColor: 'red',
    borderRadius: '10px',
    '&:hover': {
      backgroundColor: 'red',
    },
  },
   unactiveChatItem: {
    borderRadius: '10px',
    '&:hover': {
      backgroundColor: '#8c8c8c',
    },
  },
};

function Chat() {
  const [messages, setMessages] = useState([]);
  const [inputValue, setInputValue] = useState('');
  const [currentUser, setCurrentUser] = useState('You');
  const [selectedLanguage, setSelectedLanguage] = useState('english');
  const [languageAnchorEl, setLanguageAnchorEl] = useState(null);
  const [isNewChat, setIsNewChat] = useState(false);

  const [chatList, setChatList] = useState(() => {
    const savedChatList = localStorage.getItem('chatList');
    return savedChatList ? JSON.parse(savedChatList) : [];
  });
  const [activeChatId, setActiveChatId] = useState(null);

  useEffect(() => {
    const storedChatList = JSON.parse(localStorage.getItem('chatList'));
    if (storedChatList) {
      setChatList(storedChatList);
    } else {
      setChatList([]);
      localStorage.setItem('chatList', JSON.stringify([]));
    }
  }, []);

  const createNewChat = () => {
    const newChatId = Math.floor(Math.random() * 1000000);
    const newChat = { id: newChatId, name: `Chat ${chatList.length + 1}` };
  
    setChatList((prevChatList) => {
      const updatedChatList = [...prevChatList, newChat];
      localStorage.setItem('chatList', JSON.stringify(updatedChatList));
      return updatedChatList;
    });
  
    setActiveChatId(newChatId);
  };
  
  const handleSend = () => {
    if (inputValue.trim()) {
      const trimmedInput = inputValue.trim();
  
      let currentChatId = activeChatId;
  
      if (activeChatId === null) {
        const newChatId = Math.floor(Math.random() * 1000000);
        const newChatName = trimmedInput.length > 15 ? `${trimmedInput.substring(0, 15)}...` : trimmedInput;
        const newChat = { id: newChatId, name: newChatName };
  
        setChatList((prevChatList) => {
          const updatedChatList = [...prevChatList, newChat];
          localStorage.setItem('chatList', JSON.stringify(updatedChatList));
          return updatedChatList;
        });
        
        setActiveChatId(newChatId);
        currentChatId = newChatId;
      }

      const newMessage = { text: trimmedInput, sender: currentUser };
      setMessages((prevMessages) => [...prevMessages, newMessage]);
      setInputValue('');
  
      fetch(`http://localhost:8080/chat/${currentChatId}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ message: trimmedInput, language: selectedLanguage }),
      })
        .then((response) => response.json())
        .then((data) => {
          setMessages((prevMessages) => [
            ...prevMessages,
            { text: data.response, sender: currentUser === 'You' ? 'Assistant' : 'You' },
          ]);
        })
        .catch((error) => console.error('Error:', error));
  
      setCurrentUser((prevUser) => (prevUser === 'You' ? 'Assistant' : 'You'));
    }
  };
  
  
  const handleKeyPress = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const handleLanguageClick = (event) => {
    setLanguageAnchorEl(event.currentTarget);
  };

  const handleLanguageClose = () => {
    setLanguageAnchorEl(null);
  };

  const handleLanguageChange = (language) => {
    setSelectedLanguage(language);
    setLanguageAnchorEl(null);
  };

  const handleNewChat = () => {
    setMessages([]); // Очищаем историю сообщений
    setActiveChatId(null); // Устанавливаем activeChatId в null, указывая что это новый чат
  };

  const fetchChatHistory = (chatId) => {
    fetch(`http://localhost:8080/chat/${chatId}`)
      .then((response) => response.json())
      .then((data) => {
        setMessages(data.map(msg => ({
          text: msg.message,
          sender: msg.is_bot ? 'Assistant' : 'You',
        })));
      })
      .catch((error) => console.error('Error fetching chat history:', error));
  };

  useEffect(() => {
    const fetchChats = async () => {
      try {
        const response = await fetch("http://localhost:8080/chat");
        if (!response.ok) {
          throw new Error("Network response was not ok " + response.statusText);
        }
        const data = await response.json();
        setChatList(data.chats);
      } catch (error) {
        console.error("Error fetching chats:", error);
      }
    };
  
    fetchChats();
  }, []);

  return (
    <Container sx={styles.chatLayout}>
        <Box sx={styles.chatListContainer}>
            <Container sx={styles.chatListHeader}>
                <img src={sikaLogo} alt="Sika Logo" style={styles.sikaLogoImage} />
                <Fab sx={styles.newChatButton} onClick={handleNewChat}>
                    <AddIcon />
                </Fab>
            </Container>

            <List sx={styles.chatList}>
            {chatList.map((chat) => (
                <ListItem button key={chat.id} 
                onClick={() => {
                    setActiveChatId(chat.id);
                    setIsNewChat(false);
                    fetchChatHistory(chat.id);
                    // setMessages([]);
                }}
                sx={activeChatId === chat.id ? styles.activeChatItem : styles.unactiveChatItem}>
                    <ListItemText primary={chat.name} />
                    <Divider />
                </ListItem>
            ))}
            </List>
        </Box>

        <Container sx={styles.chatWindow}>

        <div>
        <IconButton
            aria-controls="language-menu"
            aria-haspopup="true"
            onClick={handleLanguageClick}
            sx={styles.languageSelector}
            >
            <Flag countryCode={getCountryCode(selectedLanguage)} style={{ marginRight: '8px' }} />
            <ExpandMoreIcon />
        </IconButton>
        <Menu
          id="language-menu"
          anchorEl={languageAnchorEl}
          keepMounted
          open={Boolean(languageAnchorEl)}
          onClose={handleLanguageClose}
        >
            <MenuItem onClick={() => handleLanguageChange('english')}>
                <Flag countryCode="GB" style={{ marginRight: '8px' }} />
                English
            </MenuItem>
            <MenuItem onClick={() => handleLanguageChange('german')}>
                <Flag countryCode="DE" style={{ marginRight: '8px' }} />
                German
            </MenuItem>
            <MenuItem onClick={() => handleLanguageChange('portuguese')}>
                <Flag countryCode="PT" style={{ marginRight: '8px' }} />
                Portuguese
            </MenuItem>
            <MenuItem onClick={() => handleLanguageChange('spanish')}>
                <Flag countryCode="ES" style={{ marginRight: '8px' }} />
                Spanish
            </MenuItem>
            <MenuItem onClick={() => handleLanguageChange('chinese')}>
                <Flag countryCode="CN" style={{ marginRight: '8px' }} />
                Chinese
            </MenuItem>
        </Menu>
        </div>

        <Divider />

      <List sx={styles.messageList}>
        {/* {messages.map((message, index) => (
          <ListItem key={index} sx={styles.messageBubble(message.sender === 'You')}>
            <ListItemText primary={message.sender} secondary={message.text} />
          </ListItem>
        ))} */}
        {messages.map((message, index) => (
            <ListItem key={index} sx={styles.messageBubble(message.sender === 'You')}>
                <ListItemText primary={message.sender} secondary={message.text} />
            </ListItem>
        ))}
      </List>

      <Divider />

      <div sx={styles.inputContainer}>
        <TextField
          sx={{
            ...styles.inputField,
            '& .MuiOutlinedInput-root': {
              '&.Mui-focused fieldset': {
                borderColor: 'red',
              },
              '& fieldset': {
                borderColor: 'red',
              },
            },
          }}
          variant="outlined"
          value={inputValue}
          onChange={(e) => setInputValue(e.target.value)}
          onKeyPress={handleKeyPress}
          fullWidth
          placeholder="Write a message ..."
          InputProps={{
            endAdornment: (
              <Button sx={styles.sendButton} color="primary" onClick={handleSend}>
                <SendIcon sx={{ color: 'red' }} />
              </Button>
            ),
            style: { backgroundColor: 'rgba(255, 255, 255, 0.5)' },
          }}
        />
      </div>
    </Container>
    </Container>
  );
}

function getCountryCode(language) {
    switch (language) {
      case 'english':
        return 'GB';
      case 'german':
        return 'DE';
      case 'portuguese':
        return 'PT';
      case 'spanish':
        return 'ES';
      case 'chinese':
        return 'CN';
      default:
        return 'US';
    }
  }  

export default Chat;
