import Header from './components/Header';
import { Layout } from 'antd';
import { Content } from 'antd/es/layout/layout';

function App(props: { children: React.ReactNode }) {
  const { children } = props;
  return (
    <Layout style={{background: "#141414"}}>
      <Header />
      <Content>
        { children }
      </Content>
    </Layout>
  );
}

export default App;
