import React, { useState } from 'react';
import { Container, TextField, Button, List, ListItem, ListItemText } from '@mui/material';
import SendIcon from '@mui/icons-material/Send';

const styles = {
    chatContainer: {
      padding: '50px',
      border: '1px solid #ccc',
      borderRadius: '10px',
      background: '#ffffff', // 1. Установите белый фон для чата
      height: '100vh',
      display: 'flex',
      flexDirection: 'column',
      justifyContent: 'space-between',
      margin: '0 auto',
      maxWidth: '800px',
    },
    messageBubble: (isOwnMessage) => ({
      alignSelf: isOwnMessage ? 'flex-end' : 'flex-start',
      maxWidth: '50%',
      padding: '10px 20px',
      borderRadius: '20px',
      margin: '10px',
      backgroundColor: '#f5f5f5',
    }), 
    inputContainer: {
      display: 'flex',
      padding: '10px',
      backgroundColor: '#f0f0f0', // 4. Обновите цвет фона для контейнера ввода
      borderRadius: '10px', 
      position: 'relative',
    },
    inputField: {
      flexGrow: 1,  
      borderRadius: '10px',
      overflow: 'hidden',
      marginRight: '10px',
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
    }
  };


function Chat() {
    const [messages, setMessages] = useState([]);
    const [inputValue, setInputValue] = useState('');
    const [currentUser, setCurrentUser] = useState('You');

    const chatID = 111

    const handleSend = () => {
        if (inputValue.trim()) {
            const newMessage = { text: inputValue, sender: currentUser };
            setMessages((prevMessages) => [...prevMessages, newMessage]);
            setInputValue('');
            
            fetch(`http://localhost:8080/chat/${chatID}`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ message: inputValue }),
            })
            .then(response => response.json())
            .then(data => {
                setMessages((prevMessages) => [...prevMessages, { text: data.response, sender: currentUser === 'You' ? 'Friend' : 'You' }]);
            })
            .catch(error => console.error('Error:', error));
            
            setCurrentUser((prevUser) => (prevUser === 'You' ? 'Friend' : 'You'));
        }
    };

    const handleKeyPress = (e) => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            handleSend();
        }
    };

    return (
      <Container sx={styles.chatContainer}>
        <List>
        {messages.map((message, index) => (
            <ListItem 
                key={index} 
                sx={styles.messageBubble(message.sender === 'You')}>
                <ListItemText primary={message.sender} secondary={message.text} />
            </ListItem>
            ))}


        </List>
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
            placeholder="Write a message..." 
            InputProps={{
                endAdornment: (
                <Button
                    sx={styles.sendButton}
                    color="primary"
                    onClick={handleSend}
                >
                    <SendIcon sx={{ color: 'red' }} /> 
                </Button>
                ),
                style: { backgroundColor: 'rgba(255, 255, 255, 0.5)' },
            }}
            />
         </div>
      </Container>
    );
}

export default Chat;
  