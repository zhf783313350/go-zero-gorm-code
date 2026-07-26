import asyncio
import base64
import os

from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession, async_sessionmaker
from sqlalchemy import text
from cryptography.hazmat.primitives.kdf.scrypt import Scrypt
from cryptography.hazmat.backends import default_backend

DEFAULT_PLAIN_PASSWORD = "123456"
DATABASE_URL = "postgresql+asyncpg://postgres:123456@localhost:5432/zero_gorm_code"

engine = create_async_engine(DATABASE_URL)
AsyncSessionLocal = async_sessionmaker(bind=engine, expire_on_commit=False, class_=AsyncSession)

def hash_password(password: str) -> str:
    salt = os.urandom(16)
    kdf = Scrypt(
        salt=salt,
        length=32,
        n=1 << 14,
        r=8,
        p=1,
        backend=default_backend()
    )
    hash_bytes = kdf.derive(password.encode('utf-8'))
    result = salt + hash_bytes
    return base64.b64encode(result).decode('utf-8')

async def main():
    async with AsyncSessionLocal() as db:
        try:
            hashed_password = hash_password(DEFAULT_PLAIN_PASSWORD)
            print(f"🔑 明文密码: '{DEFAULT_PLAIN_PASSWORD}'")
            print(f"🔒 scrypt 哈希密文: '{hashed_password}'")
            
            await db.execute(
                text("UPDATE users SET password = :hp WHERE password IS NOT NULL"),
                {"hp": hashed_password}
            )
            
            await db.commit()
            print("✅ 数据库密码迁移成功！")
        except Exception as e:
            await db.rollback()
            print(f"❌ 迁移失败: {e}")

if __name__ == "__main__":
    asyncio.run(main())